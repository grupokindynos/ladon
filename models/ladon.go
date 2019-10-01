package models

type PrepareVoucher struct {
	Coin           string `json:"coin"`
	VoucherType    int    `json:"voucher_type"`
	VoucherVariant string `json:"voucher_variant"`
}

type PrepareVoucherResponse struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type PrepareVoucherInfo struct {
	Coin           string  `json:"coin"`
	VoucherType    int     `json:"voucher_type"`
	VoucherVariant string  `json:"voucher_variant"`
	Address        string  `json:"address"`
	Amount         float64 `json:"amount"`
	Timestamp      int64   `json:"timestamp"`
	FiatAmount     float64 `json:"fiat_amount"`
	VoucherName    string  `json:"voucher_name"`
}

type RedeemCodeVoucher struct {
	VoucherID  string `json:"voucher_id"`
	RedeemCode string `json:"redeem_code"`
}
