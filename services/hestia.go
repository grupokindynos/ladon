package services

import (
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/jws"
	"os"
)

var HestiaURL = hestia.ProductionURL

func StoreNewVoucher() {

}

func GetVoucherInfo() {

}

func ValidateToken(token string) bool {
	encodedToken, err := jws.EncodeJWS(token, os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return false
	}
	return hestia.VerifyToken(encodedToken, "ladon")
}
