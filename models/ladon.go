package models

type PrepareVoucher struct {
	Coin           string `json:"coin"`
	VoucherType    int    `json:"voucher_type"`
	VoucherVariant string `json:"voucher_variant"`
	Country        string `json:"country"`
}

type PrepareVoucherResponse struct {
	Address string  `json:"address"`
	Amount  int32 `json:"amount"`
}

type PrepareVoucherInfo struct {
	Coin           string  `json:"coin"`
	VoucherType    int     `json:"voucher_type"`
	VoucherVariant string  `json:"voucher_variant"`
	Address        string  `json:"address"`
	Amount         int32 `json:"amount"`
	Timestamp      int64   `json:"timestamp"`
	FiatAmount     int32 `json:"fiat_amount"`
	VoucherName    string  `json:"voucher_name"`
}

type RedeemCodeVoucher struct {
	VoucherID  string `json:"voucher_id"`
	RedeemCode string `json:"redeem_code"`
}

type VouchersList struct {
	Austria      []Voucher `json:"austria"`
	Belgium      []Voucher `json:"belgium"`
	Canada       []Voucher `json:"canada"`
	Croatia      []Voucher `json:"croatia"`
	Cyprus       []Voucher `json:"cyprus"`
	Czechia      []Voucher `json:"czechia"`
	Denmark      []Voucher `json:"denmark"`
	Estonia      []Voucher `json:"estonia"`
	Finland      []Voucher `json:"finland"`
	France       []Voucher `json:"france"`
	Germany      []Voucher `json:"germany"`
	GreatBritain []Voucher `json:"great_britain"`
	Greece       []Voucher `json:"greece"`
	Hungary      []Voucher `json:"hungary"`
	Ireland      []Voucher `json:"ireland"`
	Italy        []Voucher `json:"italy"`
	Lichtenstein []Voucher `json:"lichtenstein"`
	Luxembourg   []Voucher `json:"luxembourg"`
	Malta        []Voucher `json:"malta"`
	Netherland   []Voucher `json:"netherland"`
	Norway       []Voucher `json:"norway"`
	Poland       []Voucher `json:"poland"`
	Portugal     []Voucher `json:"portugal"`
	Slovakia     []Voucher `json:"slovakia"`
	Slovenia     []Voucher `json:"slovenia"`
	Spain        []Voucher `json:"spain"`
	Sweden       []Voucher `json:"sweden"`
	Switzerland  []Voucher `json:"switzerland"`
	Turkey       []Voucher `json:"turkey"`
	Usa          []Voucher `json:"usa"`
}
