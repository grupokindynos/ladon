package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/ppat"
	"github.com/grupokindynos/ladon/controllers"
	"github.com/grupokindynos/ladon/processor"
	"github.com/grupokindynos/ladon/services"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"sync"
	"time"
)

type CurrentTime struct {
	Hour   int
	Day    int
	Minute int
	Second int
}

var currTime CurrentTime

func init() {
	_ = godotenv.Load()
}

func main() {

	currTime = CurrentTime{
		Hour:   time.Now().Hour(),
		Day:    time.Now().Day(),
		Minute: time.Now().Minute(),
		Second: time.Now().Second(),
	}

	go timer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	App := GetApp()
	_ = App.Run(":" + port)
}

func GetApp() *gin.Engine {
	App := gin.Default()
	App.Use(cors.Default())
	ApplyRoutes(App)
	return App
}

func ApplyRoutes(r *gin.Engine) {
	api := r.Group("/")
	{
		bitcouService := services.InitService()
		vouchersCtrl := controllers.VouchersController{BitcouService: bitcouService}
		// New voucher routes
		api.POST("/prepare", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetToken) })
		api.POST("/new", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.Store) })
		// Service status
		api.GET("/status", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetServiceStatus) })
		// Available vouchers
		api.GET("/list", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetList) })
		// Processing voucher information
		api.GET("/info/:voucherid", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetInfo) })
	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}

func ValidateRequest(c *gin.Context, method func(payload []byte, uid string, voucherid string) (interface{}, error)) {
	token := c.GetHeader("token")
	voucherid := c.Param("voucherid")
	if token == "" {
		responses.GlobalResponseNoAuth(c)
		return
	}
	var body []byte
	err := c.BindJSON(&body)
	if err != nil {
		responses.GlobalResponseNoAuth(c)
		return
	}
	valid, payload, uid := ppat.VerifyPPATToken("ladon", os.Getenv("MASTER_PASSWORD"), token, body, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"), os.Getenv("HESTIA_PUBLIC_KEY"))
	if !valid {
		responses.GlobalResponseNoAuth(c)
		return
	}
	res, err := method(payload, uid, voucherid)
	responses.GlobalResponseError(res, err, c)
}

func timer() {
	for {
		time.Sleep(1 * time.Second)
		currTime = CurrentTime{
			Hour:   time.Now().Hour(),
			Day:    time.Now().Day(),
			Minute: time.Now().Minute(),
			Second: time.Now().Second(),
		}
		if currTime.Second == 0 {
			var wg sync.WaitGroup
			wg.Add(1)
			runCrons(&wg)
			wg.Wait()
		}
	}
}

func runCrons(mainWg *sync.WaitGroup) {
	defer func() {
		mainWg.Done()
	}()
	var wg sync.WaitGroup
	wg.Add(8)
	go runCronMinutes(1, processor.Start, &wg) // 1 minute
	wg.Wait()
}

func runCronMinutes(schedule int, function func(), wg *sync.WaitGroup) {
	go func() {
		defer func() {
			wg.Done()
		}()
		remainder := currTime.Minute % schedule
		if remainder == 0 {
			function()
		}
		return
	}()

}
