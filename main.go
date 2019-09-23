package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/ladon/config"
	"github.com/grupokindynos/ladon/controllers"
	"github.com/grupokindynos/ladon/processor"
	"github.com/grupokindynos/ladon/services"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strings"
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
		api.GET("/list", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetVouchersList) })
	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}

func ValidateRequest(c *gin.Context, method func() (interface{}, error)) {
	reqToken, ok := c.Request.Header["Authorization"]
	if !ok {
		config.GlobalResponseNoAuth(c)
		return
	}
	splitToken := strings.Split(reqToken[0], "Bearer ")
	token := splitToken[1]
	// If there is no token on the header, return non-authed
	if token == "" {
		config.GlobalResponseNoAuth(c)
		return
	}
	valid := services.ValidateToken(token)
	if !valid {
		config.GlobalResponseNoAuth(c)
		return
	}
	res, err := method()
	config.GlobalResponseError(res, err, c)
}

func timer() {
	for {
		time.Sleep(time.Duration(1 * time.Second))
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
