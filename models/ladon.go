package models

import "github.com/grupokindynos/adrestia-go/models"

type PrepareVoucher struct {
	Coin           string `json:"coin"`
	VoucherType    int    `json:"voucher_type"`
	VoucherVariant string `json:"voucher_variant"`
	Country        string `json:"country"`
	VoucherName    string `json:"name"`
	VoucherImage   string `json:"image"`
	PhoneNumber    string `json:"phone_nb"`
	ProviderId     int32  `json:"provider_id"`
	Valid          int32  `json:"valid"`
}

type PrepareVoucherResponse struct {
	Payment PaymentInfo `json:"payment"`
	Fee     PaymentInfo `json:"fee"`
}

type PaymentInfo struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type PrepareVoucherInfo struct {
	ID               string      `json:"id"`
	Coin             string      `json:"coin"`
	VoucherType      int         `json:"voucher_type"`
	VoucherVariant   string      `json:"voucher_variant"`
	Payment          PaymentInfo `json:"payment"`
	FeePayment       PaymentInfo `json:"fee_payment"`
	BitcouPayment    PaymentInfo `json:"bitcou_payment"`
	BitcouFeePayment PaymentInfo `json:"bitcou_fee_payment"`
	BitcouID         string      `json:"bitcou_id"`
	Timestamp        int64       `json:"timestamp"`
	AmountEuro       string      `json:"amount_euro"`
	AmountFeeEuro    string      `json:"amount_fee_euro"`
	Name             string      `json:"name"`
	Image            string      `json:"image"`
	PhoneNumber      int64       `json:"phone_nb"`
	ProviderId       int32       `json:"provider_id"`
	Valid            int32       `json:"valid"`
}

type PrepareVoucherInfoV2 struct {
	ID             string                     `json:"id"`
	Timestamp      int64                      `json:"timestamp"`
	AmountEuro     int64                      `json:"amount_euro"`
	UserPayment    PaymentInfo                `json:"user_payment"`
	Coin           string                     `json:"coin"`
	VoucherType    int                        `json:"voucher_type"`
	VoucherVariant int                        `json:"voucher_variant"`
	Name           string                     `json:"name"`
	PhoneNumber    int64                      `json:"phone_nb"`
	ProviderId     string                     `json:"provider_id"`
	Path           models.VoucherPathResponse `json:"path"`
}

type StoreVoucher struct {
	RawTx         string `json:"raw_tx"`
	FeeTx         string `json:"fee_tx"`
	RefundAddr    string `json:"refund_addr"`
	RefundFeeAddr string `json:"refund_fee_addr"`
}

type RedeemCodeVoucher struct {
	VoucherID  string `json:"voucher_id"`
	RedeemCode string `json:"redeem_code"`
}
