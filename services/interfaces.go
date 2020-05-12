package services

import (
	amodels "github.com/grupokindynos/adrestia-go/models"
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
	GetVoucherInfoV2(voucherid string) (hestia.VoucherV2, error)
	UpdateVoucher(voucherData hestia.Voucher) (string, error)
	UpdateVoucherV2(voucherData hestia.VoucherV2) (string, error)
	GetVouchersByStatusV2(status hestia.VoucherStatusV2) ([]hestia.VoucherV2, error)
	GetVouchersByTimestamp(uid string, timestamp string) (vouchers []hestia.Voucher, err error)
	GetVouchersByTimestampV2(uid string, timestamp string) (vouchers []hestia.VoucherV2, err error)
}

type BitcouService interface {
	GetPhoneTopUpList(phoneNb string) ([]int, error)
	GetTransactionInformation(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponse, error)
	GetTransactionInformationV2(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponseV2, error)
}

type AdrestiaService interface {
	DepositInfo(depositParams amodels.DepositParams) (depositInfo amodels.DepositInfo, err error)
	Trade(tradeParams hestia.Trade) (txId string, err error)
	GetTradeStatus(tradeParams hestia.Trade) (tradeInfo hestia.ExchangeOrderInfo, err error)
	Withdraw(withdrawParams amodels.WithdrawParams) (withdrawal amodels.WithdrawInfo, err error)
	GetWithdrawalTxHash(withdrawParams amodels.WithdrawInfo) (txId string, err error)
	GetPath(fromCoin string) (path amodels.VoucherPathResponse, err error)
}
