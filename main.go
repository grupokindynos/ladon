package main

import (
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/ppat"
	"github.com/grupokindynos/ladon/controllers"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/processor"
	"github.com/grupokindynos/ladon/services"
	"github.com/joho/godotenv"
	"log"
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

var (
	currTime           CurrentTime
	prepareVouchersMap = make(map[string]models.PrepareVoucherInfo)
)

const prepareVoucherTimeframe = 60 * 5 // 5 minutes

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
	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = true
	corsConf.AllowHeaders = []string{"token", "service", "content-type"}
	App.Use(cors.New(corsConf))
	ApplyRoutes(App)
	return App
}

func ApplyRoutes(r *gin.Engine) {
	bitcouService := services.InitService()
	vouchersCtrl := &controllers.VouchersController{BitcouService: bitcouService, PreparesVouchers: prepareVouchersMap}
	go checkAndRemoveVouchers(vouchersCtrl)
	api := r.Group("/")
	{
		// New voucher routes
		api.POST("/prepare", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetToken) })
		api.POST("/new", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.Store) })
		// Service status
		api.GET("/status", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetServiceStatus) })
		// Available vouchers
		api.GET("/list", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetList) })
		api.GET("/phone/:phone", func(context *gin.Context) { ValidateRequest(context, vouchersCtrl.GetListForPhone) })

		// Bitcou endpoint for a voucher redeem
		api.POST("/redeem", vouchersCtrl.Update)
	}
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})
}

func ValidateRequest(c *gin.Context, method func(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error)) {
	fbToken := c.GetHeader("token")
	voucherid := c.Param("voucherid")
	phoneNb := c.Param("phone")
	if fbToken == "" {
		responses.GlobalResponseNoAuth(c)
		return
	}
	tokenBytes, _ := c.GetRawData()
	var ReqBody hestia.BodyReq
	if len(tokenBytes) > 0 {
		err := json.Unmarshal(tokenBytes, &ReqBody)
		if err != nil {
			responses.GlobalResponseError(nil, err, c)
			return
		}
	}
	valid, payload, uid, err := ppat.VerifyPPATToken("https://hestia.polispay.com", "ladon", os.Getenv("MASTER_PASSWORD"), fbToken, ReqBody.Payload, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"), os.Getenv("HESTIA_PUBLIC_KEY"))
	if !valid {
		responses.GlobalResponseNoAuth(c)
		return
	}
	response, err := method(payload, uid, voucherid, phoneNb)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	encrypted, err := jwt.EncryptJWE(uid, response)
	responses.GlobalResponseError(encrypted, err, c)
	return
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
	wg.Add(1)
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

func checkAndRemoveVouchers(ctrl *controllers.VouchersController) {
	for {
		time.Sleep(time.Second * 60)
		log.Print("Removing obsolete vouchers request")
		count := 0
		for k, v := range ctrl.PreparesVouchers {
			if time.Now().Unix() > v.Timestamp+prepareVoucherTimeframe {
				count += 1
				ctrl.RemoveVoucherFromMap(k)
			}
		}
		log.Printf("Removed %v vouchers", count)
	}
}
