package processor

import (
	"github.com/grupokindynos/ladon/services"
)

type BitcouPayment struct {
	Hestia services.HestiaService
	Adrestia services.AdrestiaService
}