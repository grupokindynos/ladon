package controllers

import (
	"encoding/json"
	"errors"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	commonErrors "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/explorer"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type VouchersControllerV2 struct {
	PreparesVouchers map[string]models.PrepareVoucherInfoV2
	mapLock          sync.RWMutex
	TxsAvailable     bool
	Plutus           services.PlutusService
	Hestia           services.HestiaService
	HestiaTest		 services.HestiaService
	Bitcou           services.BitcouService
	Obol             obol.ObolService
	Adrestia         services.AdrestiaService
}

var (
	Whitelist = make(map[string]bool)
)

func (vc *VouchersControllerV2) StatusV2(_ []byte, uid string, _ string, _ string, test bool) (interface{}, error) {
	if test {
		return true, nil
	}
	Whitelist["gwY3fy79LZMtUbSNBDoom7llGfh2"] = true
	Whitelist["yEF8YP4Ou9aCEqSPQPqDslviGfT2"] = true
	Whitelist["TO3FrEneQcf2RN2QdL8paY6IvBF2"] = true
	Whitelist["YIrr2a42lcZi9djePQH7OrLbGzs1"] = true
	Whitelist["Egc6XKdkmigtWzuyq0YordjWODq1"] = true
	Whitelist["HMOXcoZJxfMKFca9IukZIaqI2Z02"] = true
	Whitelist["fENeiyOGJURJK9qielqR7OrxciJ3"] = true

	if val, ok := Whitelist[uid]; ok {
		return val, nil
	}
	status, err := vc.Hestia.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	return status.Vouchers.Service, nil
}

func (vc *VouchersControllerV2) PrepareV2(payload []byte, uid string, voucherid string, phoneNb string, test bool) (interface{}, error) {
	hestiaDb := vc.getHestiaDb(test)
	config, err := vc.StatusV2(payload, uid, voucherid, phoneNb, test)
	if err != nil {
		return nil, err
	}
	service := cast.ToBool(config)
	if !service {
		return nil, err
	}
	// Grab information on the payload
	var PrepareVoucher models.PrepareVoucher
	err = json.Unmarshal(payload, &PrepareVoucher)
	if err != nil {
		return nil, err
	}
	coinsConfig, err := hestiaDb.GetCoinsConfig()
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
		return nil, errors.New("coin not available")
	}
	// Create a VoucherID
	newVoucherID := utils.RandomString()

	//get email
	email, err := hestiaDb.GetUserInfo(uid)
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

	voucherInfo, err := hestiaDb.GetVoucherInfoV2(PrepareVoucher.Country, strconv.Itoa(int(PrepareVoucher.ProductId)))
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
	PrepareVoucher.Valid = int32(voucherInfo.Valid)

	// Exchange Path
	pathInfo, err := vc.Adrestia.GetPath(PrepareVoucher.Coin, voucherInfo.Variants[variantIndex].Price / 100)
	if err != nil {
		if err == commonErrors.ErrorNotSupportedAmount {
			return nil, err
		}

		err = commonErrors.ErrorFillingPaymentInformation
		return nil, err
	}

	paymentAddr := pathInfo.Address
	feePaymentAddr := pathInfo.Address

	// TODO VALIDATE PRICE IS ALWAYS IN EURO
	euroRate, err := vc.Obol.GetCoin2FIATRate(PrepareVoucher.Coin, "EUR")
	if err != nil {
		return nil, err
	}

	var purchaseAmountEuro float64
	purchaseAmountEuro = voucherInfo.Variants[variantIndex].Price / 100

	// purchaseAmountEuro = 1.1


	balance, err := vc.Bitcou.GetAccountBalanceV2()
	if err != nil {
		return nil, err
	}
	if purchaseAmountEuro > float64(balance.Amount)/100 && !test {
		log.Println("not enough balance in floating account to fulfill request")
		return nil, commonErrors.ErrorNotEnoughDash // TODO RENAME ERROR
	}

	if 100 > float64(balance.Amount) / 100 {
		log.Println("balance below 100 EURO please refill")
		// TODO Configure telegram alert || automatic refill
	}

	securityFactor := getSecurityFactor(strings.ToUpper(PrepareVoucher.Coin))
	// Amounts for amount and fees in float representation
	paymentAmount := decimal.NewFromFloat(purchaseAmountEuro / (euroRate * securityFactor))
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
	vouchers, err := hestiaDb.GetVouchersByTimestampV2(uid, timestamp)
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

	if totalAmountEuro > 210.0 && !test {
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
		Country: PrepareVoucher.Country,
	}

	vc.AddVoucherToMapV2(uid, prepareVoucher)
	return res, nil
}

func getSecurityFactor(coin string) float64 {
	exCoins := [3]string{"RPD", "TELOS", "FYD"}
	medCoins := [2]string{"POLIS", "MW"}
	exHighCoins := [6]string{"IDX", "BITG", "NULS", "CRW", "COLX", "GTH"}
	for _, c := range exCoins {
		if c == coin {
			return 0.982
		}
	}
	for _, c := range medCoins {
		if c == coin {
			return 0.935
		}
	}
	for _, c := range exHighCoins {
		if c == coin {
			return 0.90
		}
	}
	return 1.0
}

func (vc *VouchersControllerV2) StoreV2(payload []byte, uid string, voucherId string, _ string, test bool) (interface{}, error) {
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
		Country: storedVoucher.Country,
	}

	vc.RemoveVoucherFromMapV2(uid)

	hestiaDb := vc.getHestiaDb(test)
	voucherId, err = hestiaDb.UpdateVoucherV2(voucher)
	if err != nil {
		return nil, err
	}

	if !test {
		go vc.decodeAndCheckTxV2(voucher, storedVoucher, voucherPayments.RawTx)
	}
	return voucherId, nil
}

func (vc *VouchersControllerV2) GetListForPhoneV2(_ []byte, _ string, _ string, phoneNb string, _ bool) (interface{}, error) {
	vouchersAvailable, err := vc.Bitcou.GetPhoneTopUpListV2(phoneNb)
	if err != nil {
		return nil, err
	}
	return vouchersAvailable, nil
}

func (vc *VouchersControllerV2) decodeAndCheckTxV2(voucherData hestia.VoucherV2, _ models.PrepareVoucherInfoV2, rawTx string) {
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
		_, err = hestiaDb.UpdateVoucherV2(voucherData)
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
	explorerWrapper, _ := explorer.NewExplorerFactory().GetExplorerByCoin(*coinConfig)
	return explorerWrapper.SendTxWithMessage(rawTx)
}

func (vc *VouchersControllerV2) getHestiaDb(test bool) services.HestiaService {
	if test {
		return vc.HestiaTest
	}

	return vc.Hestia
}
