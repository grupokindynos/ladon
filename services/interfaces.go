package services

import (
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/ladon/models"
)

type PlutusService interface {
	GetNewPaymentAddress(coin string) (addr string, err error)
	SubmitPayment(body plutus.SendAddressBodyReq) (txid string, err error)
	ValidateRawTx(body plutus.ValidateRawTxReq) (valid bool, err error)
	GetWalletBalance(coin string) (plutus.Balance, error)
}

type HestiaService interface {
	GetVouchersStatus() (hestia.Config, error)
	GetCoinsConfig() ([]hestia.Coin, error)
	GetVoucherInfo(voucherid string) (hestia.Voucher, error)
	UpdateVoucher(voucherData hestia.Voucher) (string, error)
}

type BitcouService interface {
	GetPhoneTopUpList(phoneNb string) ([]int, error)
	GetTransactionInformation(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponse, error)
}
