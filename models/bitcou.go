package models

type Countries struct {
	Austria      bool `json:"austria"`
	Belgium      bool `json:"belgium"`
	Bulgaria     bool `json:"bulgaria"`
	Canada       bool `json:"canada"`
	Croatia      bool `json:"croatia"`
	Cyprus       bool `json:"cyprus"`
	Czechia      bool `json:"czechia"`
	Denmark      bool `json:"denmark"`
	Estonia      bool `json:"estonia"`
	Finland      bool `json:"finland"`
	France       bool `json:"france"`
	Germany      bool `json:"germany"`
	GreatBritain bool `json:"great_britain"`
	Greece       bool `json:"greece"`
	Hongkong     bool `json:"hongkong"`
	Hungary      bool `json:"hungary"`
	Indonesia    bool `json:"indonesia"`
	Ireland      bool `json:"ireland"`
	Italy        bool `json:"italy"`
	Lichtenstein bool `json:"lichtenstein"`
	Luxembourg   bool `json:"luxembourg"`
	Malaysia     bool `json:"malaysia"`
	Malta        bool `json:"malta"`
	Mexico       bool `json:"mexico"`
	Netherland   bool `json:"netherland"`
	Norway       bool `json:"norway"`
	Poland       bool `json:"poland"`
	Portugal     bool `json:"portugal"`
	Russia       bool `json:"russia"`
	Singapore    bool `json:"singapore"`
	Slovakia     bool `json:"slovakia"`
	Slovenia     bool `json:"slovenia"`
	Spain        bool `json:"spain"`
	Sweden       bool `json:"sweden"`
	Switzerland  bool `json:"switzerland"`
	Turkey       bool `json:"turkey"`
	Usa          bool `json:"usa"`
}

type RedeemPlace struct {
	Market bool `json:"market"`
	Online bool `json:"online"`
}

type Shipping struct {
	EMail bool `json:"e_mail"`
	Mail  bool `json:"mail"`
	Print bool `json:"print"`
}

type Benefits struct {
	Data           bool `json:"Data"`
	DigitalProduct bool `json:"DigitalProduct"`
	Gaming         bool `json:"Gaming"`
	Giftcards      bool `json:"Giftcards"`
	Minutes        bool `json:"Minutes"`
	Mobile         bool `json:"Mobile"`
	Phone          bool `json:"Phone"`
	TV             bool `json:"TV"`
	Utility        bool `json:"Utility"`
}

type Variants struct {
	Currency  string  `json:"currency"`
	Ean       string  `json:"ean"`
	Price     float64 `json:"price"`
	Value     float64 `json:"value"`
	VariantID string  `json:"variant_id"`
}

type Voucher struct {
	Countries   Countries   `json:"countries"`
	Image       string      `json:"image"`
	Name        string      `json:"name"`
	ProductID   int         `json:"product_id"`
	RedeemPlace RedeemPlace `json:"redeem_place"`
	Shipping    Shipping    `json:"shipping"`
	TraderID    int         `json:"trader_id"`
	Variants    []Variants  `json:"variants"`
	Benefits    Benefits    `json:"benefits"`
}

type MetaData struct {
	Datetime string `json:"datetime"`
}

type PurchaseInfo struct {
	TransactionID string `json:"transaction_id"`
	ProductID     int32  `json:"product_id"`
	VariantID     int32  `json:"variant_id"`
	PhoneNB       int64  `json:"phone_nb"`
	Email         string `json:"email"`
	KYC  		  bool   `json:"kyc"`
}

type BitcouBaseResponse struct {
	Data []interface{} `json:"data"`
	Meta MetaData      `json:"meta"`
}

type PurchaseInfoResponseV2 struct {
	TxId string        `json:"txn_id"`
	AmountEuro string `json:"amount_euro"`
	RedeemData string  `json:"redeem_data"`
}

type PurchaseInfoResponse struct {
	Amount              float64 `json:"amount"`
	AmountEuro          string  `json:"amount_euro"`
	BitcouTransactionID string  `json:"txn_id"`
	Address             string  `json:"address"`
	Timeout             int64   `json:"timeout"`
	QRCode              string  `json:"qr_code"`
}

type BitcouPhoneResponseList struct {
	Meta struct {
		Datetime string `json:"datetime"`
	} `json:"meta"`
	Data []struct {
		ProductID int `json:"product_id"`
	} `json:"data"`
}

type BitcouPhoneBodyReq struct {
	PhoneNumber string `json:"phone_number"`
}
