package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	commonErrors "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"github.com/shopspring/decimal"
	"log"
	"strconv"
	"sync"
	"time"
)

type VouchersControllerV2 struct {
	PreparesVouchers map[string]models.PrepareVoucherInfoV2
	mapLock          sync.RWMutex
	TxsAvailable     bool
	Plutus           services.PlutusService
	Hestia           services.HestiaService
	Bitcou           services.BitcouService
	Obol             obol.ObolService
	Adrestia         services.AdrestiaService
}

func (vc *VouchersControllerV2) PrepareV2(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	/*config, err := vc.Hestia.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	if !config.Vouchers.Service {
		return nil, err
	}*/
	// Grab information on the payload
	var PrepareVoucher models.PrepareVoucher
	err := json.Unmarshal(payload, &PrepareVoucher)
	if err != nil {
		return nil, err
	}
	coinsConfig, err := vc.Hestia.GetCoinsConfig()
	if err != nil {
		return nil, err
	}
	var paymentCoinConfig hestia.Coin
	for _, coin := range coinsConfig {
		if coin.Ticker == PrepareVoucher.Coin {
			paymentCoinConfig = coin
		}
	}
	if paymentCoinConfig.Ticker == "" || !paymentCoinConfig.Vouchers.Available {
		return nil, err
	}
	// Create a VoucherID
	newVoucherID := utils.RandomString()
	// Get a payment address from the hot-wallets
	// exchange path
	pathInfo, err := vc.Adrestia.GetPath(PrepareVoucher.Coin)
	if err != nil {
		err = commonErrors.ErrorFillingPaymentInformation
		return nil, err
	}
	paymentAddr := pathInfo.Address
	feePaymentAddr := pathInfo.Address

	//get email
	email, err := vc.Hestia.GetUserInfo(uid)
	if err != nil {
		return nil, err
	}

	// Convert the variant id to int
	voucherVariantInt, _ := strconv.Atoi(PrepareVoucher.VoucherVariant)
	providerIdInt := strconv.Itoa(int(PrepareVoucher.ProviderId))
	if PrepareVoucher.PhoneNumber == "" {
		PrepareVoucher.PhoneNumber = "0"
	}
	phoneNumber, err := strconv.Atoi(PrepareVoucher.PhoneNumber)
	if err != nil {
		return nil, err
	}

	voucherInfo, err := vc.Hestia.GetVoucherInfoV2(PrepareVoucher.Country, strconv.Itoa(int(PrepareVoucher.ProductId)))
	if err != nil {
		return nil, err
	}

	variantIndex := -1
	for i, variant := range voucherInfo.Variants {
		if PrepareVoucher.VoucherVariant == variant.VariantID {
			variantIndex = i
			break
		}
	}
	fmt.Println(variantIndex)
	PrepareVoucher.Valid = int32(voucherInfo.Valid)

	// TODO VALIDATE PRICE IS ALWAYS IN EURO
	euroRate, err := vc.Obol.GetCoin2FIATRate(PrepareVoucher.Coin, "EUR")
	if err != nil {
		return nil, err
	}

	purchaseAmountEuro := float64(110) / 100
	// purchaseAmountEuro := voucherInfo.Variants[variantIndex].Price / 100 // TODO This is for production

	// Amounts for amount and fees in float representation
	paymentAmount := decimal.NewFromFloat(purchaseAmountEuro / euroRate)
	feePercentage := paymentCoinConfig.Vouchers.FeePercentage / float64(100)
	feeAmount := paymentAmount.Mul(decimal.NewFromFloat(feePercentage))
	// check if its a token to adjust to the amount
	coinConfig, err := coinfactory.GetCoin(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
		aux, _ := paymentAmount.Float64()
		paymentAmount = decimal.NewFromFloat(roundTo(aux, coinConfig.Info.Decimals))
	}

	paymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: paymentAmount.Mul(decimal.NewFromInt(1e8)).IntPart()}
	feeInfo := models.PaymentInfo{Address: feePaymentAddr, Amount: feeAmount.Mul(decimal.NewFromInt(1e8)).IntPart()}
	userPaymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: paymentInfo.Amount + feeInfo.Amount}

	// Validate that users hasn't bought more than 210 euro in vouchers on the last 24 hours.
	timestamp := strconv.FormatInt(time.Now().Unix()-24*3600, 10)
	vouchers, err := vc.Hestia.GetVouchersByTimestampV2(uid, timestamp)
	if err != nil {
		return nil, err
	}

	totalAmountEuro := purchaseAmountEuro

	for _, voucher := range vouchers {
		if voucher.Status != hestia.VoucherStatusV2Error && voucher.Status != hestia.VoucherStatusV2Refunded {
			amEr := float64(voucher.AmountEuro)
			amEr /= 100
			totalAmountEuro += amEr
		}
	}

	if totalAmountEuro > 210.0 {
		return nil, commonErrors.ErrorVoucherLimit
	}

	// Build the response
	res := models.PrepareVoucherResponse{
		Payment: paymentInfo,
		Fee:     feeInfo,
	}
	// Store on local cache
	prepareVoucher := models.PrepareVoucherInfoV2{
		ID:             newVoucherID,
		Coin:           PrepareVoucher.Coin,
		Timestamp:      time.Now().Unix(),
		VoucherType:    PrepareVoucher.VoucherType,
		VoucherVariant: voucherVariantInt,
		Path:           pathInfo,
		UserPayment:    userPaymentInfo,
		AmountEuro:     110,
		Name:           PrepareVoucher.VoucherName,
		PhoneNumber:    int64(phoneNumber),
		ProviderId:     providerIdInt,
		Email:          email,
		ShippingMethod: voucherInfo.Shipping.GetEnum(),
		Valid: PrepareVoucher.Valid,
	}

	vc.AddVoucherToMapV2(uid, prepareVoucher)
	return res, nil
}

func (vc *VouchersControllerV2) StoreV2(payload []byte, uid string, voucherId string, phoneNb string) (interface{}, error) {
	var voucherPayments models.StoreVoucher
	err := json.Unmarshal(payload, &voucherPayments)
	if err != nil {
		return nil, err
	}
	storedVoucher, err := vc.GetVoucherFromMapV2(uid)
	if err != nil {
		return nil, err
	}
	var status hestia.ShiftV2TradeStatus
	if len(storedVoucher.Path.InwardOrder) == 0 {
		status = hestia.ShiftV2TradeStatusCompleted
	} else {
		status = hestia.ShiftV2TradeStatusInitialized
	}
	// create trade
	exchange := ""
	var inTrade []hestia.Trade
	for _, trade := range storedVoucher.Path.InwardOrder {
		newTrade := hestia.Trade{
			OrderId:        "",
			Amount:         0,
			ReceivedAmount: 0,
			FromCoin:       trade.FromCoin,
			ToCoin:         trade.ToCoin,
			Symbol:         trade.Trade.Book,
			Side:           trade.Trade.Type,
			Exchange:       trade.Exchange,
			CreatedTime:    0,
			FulfilledTime:  0,
		}
		exchange = trade.Exchange
		inTrade = append(inTrade, newTrade)
	}
	if len(inTrade) > 0 {
		inTrade[0].Amount, _ = decimal.NewFromInt(storedVoucher.UserPayment.Amount).DivRound(decimal.NewFromInt(1e8), 8).Float64()
	}

	voucher := hestia.VoucherV2{
		Id:          storedVoucher.ID,
		CreatedTime: storedVoucher.Timestamp,
		AmountEuro:  storedVoucher.AmountEuro,
		UserPayment: hestia.Payment{
			Address:       storedVoucher.UserPayment.Address,
			Amount:        storedVoucher.UserPayment.Amount,
			Coin:          storedVoucher.Coin,
			Txid:          "",
			Confirmations: 0,
		},
		Status:        hestia.VoucherStatusV2PaymentProcessing,
		RefundAddress: voucherPayments.RefundAddr,
		VoucherId:     storedVoucher.VoucherType,
		VariantId:     storedVoucher.VoucherVariant,
		BitcouTxId:    "",
		UserId:        uid,
		RefundTxId:    "",
		FulfilledTime: 0,
		VoucherName:   storedVoucher.Name,
		PhoneNumber:   storedVoucher.PhoneNumber,
		ProviderId:    storedVoucher.ProviderId,
		RedeemCode:    "",
		Conversion: hestia.DirectionalTrade{
			Conversions:    inTrade,
			Status:         status,
			Exchange:       exchange,
			WithdrawAmount: 0.0,
		},
		Email:          storedVoucher.Email,
		ShippingMethod: storedVoucher.ShippingMethod,
		Message: "",
		Valid: storedVoucher.Valid,
	}

	vc.RemoveVoucherFromMapV2(uid)
	voucherId, err = vc.Hestia.UpdateVoucherV2(voucher)
	if err != nil {
		return nil, err
	}
	go vc.decodeAndCheckTxV2(voucher, storedVoucher, voucherPayments.RawTx)
	return voucherId, nil
}

func (vc *VouchersControllerV2) decodeAndCheckTxV2(voucherData hestia.VoucherV2, storedVoucherData models.PrepareVoucherInfoV2, rawTx string) {
	// Validate Payment RawTx
	/*body := plutus.ValidateRawTxReq{
		Coin:    voucherData.UserPayment.Coin,
		RawTx:   rawTx,
		Amount:  voucherData.UserPayment.Amount,
		Address: voucherData.UserPayment.Address,
	}*/
	/*valid, err := vc.Plutus.ValidateRawTx(body)
	if err != nil || !valid {
		// If fail and coin is POLIS mark as error
		if err != nil {
			voucherData.Message = "decode&checkTxV2:: could not validate RawTx: " + err.Error()
		} else {
			voucherData.Message = "decode&checkTxV2:: could not validate RawTx: tx not valid"
		}
		log.Println(voucherData.Message)
		voucherData.Status = hestia.VoucherStatusV2Error
		_, err = vc.Hestia.UpdateVoucherV2(voucherData)
		if err != nil {
			return
		}
		return
	}*/
	// Broadcast rawTx
	coinConfig, err := coinfactory.GetCoin(voucherData.UserPayment.Coin)
	if err != nil {
		voucherData.Message = "error getting payment coin info: " + err.Error()
		log.Println(voucherData.Message)
		voucherData.Status = hestia.VoucherStatusV2Error
		_, err = vc.Hestia.UpdateVoucherV2(voucherData)
		if err != nil {
			return
		}
		return
	}
	paymentTxid, err, _ := vc.broadCastTxV2(coinConfig, rawTx)
	if err != nil {
		voucherData.Message = "error broadcasting transaction: " + err.Error()
		log.Println(voucherData.Message)
		voucherData.Status = hestia.VoucherStatusV2Error
		_, err = vc.Hestia.UpdateVoucherV2(voucherData)
		if err != nil {
			return
		}
		return
	}
	// Update voucher model include txid.
	voucherData.UserPayment.Txid = paymentTxid
	_, err = vc.Hestia.UpdateVoucherV2(voucherData)
	if err != nil {
		return
	}
}

func (vc *VouchersControllerV2) AddVoucherToMapV2(uid string, voucherPrepare models.PrepareVoucherInfoV2) {
	vc.mapLock.Lock()
	vc.PreparesVouchers[uid] = voucherPrepare
	vc.mapLock.Unlock()
}

func (vc *VouchersControllerV2) RemoveVoucherFromMapV2(uid string) {
	vc.mapLock.Lock()
	delete(vc.PreparesVouchers, uid)
	vc.mapLock.Unlock()
}

func (vc *VouchersControllerV2) GetVoucherFromMapV2(uid string) (models.PrepareVoucherInfoV2, error) {
	vc.mapLock.Lock()
	voucher, ok := vc.PreparesVouchers[uid]
	vc.mapLock.Unlock()
	if !ok {
		return models.PrepareVoucherInfoV2{}, errors.New("voucher not found in cache map")
	}
	return voucher, nil
}

func (vc *VouchersControllerV2) broadCastTxV2(coinConfig *coins.Coin, rawTx string) (string, error, string) {
	if !vc.TxsAvailable {
		return "not published due no-txs flag", nil, ""
	}
	if coinConfig.Info.Token {
		coinConfig, _ = coinfactory.GetCoin("ETH")
	}
	blockbookWrapper := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)
	return blockbookWrapper.SendTxWithMessage(rawTx)
}
