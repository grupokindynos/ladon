package main

import (
	"fmt"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/services"
	"log"
	"testing"
)

func TestAddVoucher(t *testing.T) {
	voucher := hestia.VoucherV2{
		Id:             utils.RandomString(),
		CreatedTime:    0,
		AmountEuro:     0,
		UserPayment:    hestia.Payment{
			Address:       "PU4kqPfKaipNz2hyvEBFkmSpZbeiNpXSq5",
			Amount:        0,
			Coin:          "POLIS",
			Txid:          "7823aaff478e92c036ab168e1fee851debedb6101932eae90287d7a718632cc4",
			Confirmations: 20,
		},
		Status:         0,
		RefundAddress:  "",
		VoucherId:      0,
		VariantId:      0,
		BitcouTxId:     "",
		UserId:         "",
		RefundTxId:     "",
		FulfilledTime:  0,
		VoucherName:    "",
		PhoneNumber:    0,
		ProviderId:     "",
		RedeemCode:     "",
		Conversion:     hestia.DirectionalTrade{
			Conversions:    []hestia.Trade{},
			Status:         0,
			Exchange:       "",
			WithdrawAmount: 0,
		},
		ReceivedAmount: 0,
	}
	trade1 := hestia.Trade{
		OrderId:        "",
		Amount:         0,
		ReceivedAmount: 0,
		FromCoin:       "POLIS",
		ToCoin:         "BTC",
		Symbol:         "POLISBTC",
		Side:           "sell",
		Status:         0,
		Exchange:       "",
		CreatedTime:    0,
		FulfilledTime:  0,
	}
	trade2 := hestia.Trade{
		OrderId:        "",
		Amount:         0,
		ReceivedAmount: 0,
		FromCoin:       "BTC",
		ToCoin:         "TUSD",
		Symbol:         "BTCTUSD",
		Side:           "sell",
		Status:         0,
		Exchange:       "",
		CreatedTime:    0,
		FulfilledTime:  0,
	}

	voucher.Conversion.Conversions = append(voucher.Conversion.Conversions, trade1)
	voucher.Conversion.Conversions = append(voucher.Conversion.Conversions, trade2)

	h := services.HestiaRequests{HestiaURL:"HESTIA_LOCAL_URL"}
	_, err := h.UpdateVoucherV2(voucher)
	if err != nil {
		log.Println(err)
	}
}

func TestGetVoucherInfo(t *testing.T) {
	h := &services.HestiaRequests{HestiaURL:"HESTIA_LOCAL_URL"}
	voucherInfo, err := h.GetVoucherInfoV2("MX", "1164")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%+v\n", voucherInfo)
	log.Println(voucherInfo.Shipping.GetEnum())
}