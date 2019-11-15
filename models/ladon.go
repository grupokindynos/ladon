package models

type PrepareVoucher struct {
	Coin           string `json:"coin"`
	VoucherType    int    `json:"voucher_type"`
	VoucherVariant string `json:"voucher_variant"`
	Country        string `json:"country"`
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
	VoucherName      string      `json:"voucher_name"`
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
