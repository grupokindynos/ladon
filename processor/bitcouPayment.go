package processor

import (
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/ladon/services"
	"log"
)

type BitcouPayment struct {
	Hestia services.HestiaService
	Adrestia services.AdrestiaService
}

func (bp *BitcouPayment) Start() {
	vouchers, err := bp.Hestia.GetVouchersByStatusV2(hestia.VoucherStatusV2Complete)
	if err != nil {
		log.Println("BitcouPayment::Start::Unable to get completed vouchers")
		return
	}

	for _, voucher := range vouchers {

	}
}
