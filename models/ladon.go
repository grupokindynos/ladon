package models

type PrepareVoucher struct {
	Coin           string `json:"coin"`
	VoucherType    string `json:"voucher_type"`
	VoucherVariant string `json:"voucher_varian"`
}

type RedeemCodeVoucher struct {
	VoucherID  string `json:"voucher_id"`
	RedeemCode string `json:"redeem_code"`
}
