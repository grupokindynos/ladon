package controllers

import (
	"encoding/json"
	"errors"
	"github.com/grupokindynos/common/blockbook"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	commonErrors "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"github.com/olympus-protocol/ogen/utils/amount"
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
	// Get the vouchers percentage fee for PolisPay
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
	paymentAddr, err := vc.Plutus.GetNewPaymentAddress(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	feePaymentAddr := paymentAddr

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

	/*bitcouPrepareTx := models.PurchaseInfo{
		TransactionID: newVoucherID,
		ProductID:     int32(PrepareVoucher.VoucherType),
		VariantID:     int32(voucherVariantInt),
		PhoneNB:       int64(phoneNumber),
	}

	// get info from new bitcou voucher variant
	/*voucherDetails, err := vc.Bitcou.GetTransactionInformationV2(bitcouPrepareTx)
	if err != nil {
		return nil, err
	}*/

	euroRate, err := vc.Obol.GetCoin2FIATRate(PrepareVoucher.Coin, "EUR")
	if err != nil {
		return nil, err
	}

	//purchaseAmountEuro := voucherDetails.AmountEuro / 100
	//test
	purchaseAmountEuro := float64(110) / 100
	// amount for the voucher in the coin
	paymentAmountCoin := purchaseAmountEuro / euroRate
	paymentAmount, err := amount.NewAmount(paymentAmountCoin)
	if err != nil {
		return nil, err
	}
	// fee
	feePercentage := paymentCoinConfig.Vouchers.FeePercentage / float64(100)
	feeAmount, err := amount.NewAmount(paymentAmount.ToNormalUnit() * feePercentage)
	if err != nil {
		return nil, err
	}

	// check if its a token to adjust to the amount
	coinConfig, err := coinfactory.GetCoin(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	if coinConfig.Info.Token && coinConfig.Info.Tag != "ETH" {
		paymentAmount, err = amount.NewAmount(roundTo(paymentAmount.ToNormalUnit(), coinConfig.Info.Decimals))
		if err != nil {
			return nil, err
		}
	}

	paymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: int64(paymentAmount.ToUnit(amount.AmountSats))}
	feeInfo := models.PaymentInfo{Address: feePaymentAddr, Amount: int64(feeAmount.ToUnit(amount.AmountSats))}
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

	// exchange path
	pathInfo, err := vc.Adrestia.GetPath(PrepareVoucher.Coin)
	if err != nil {
		err = commonErrors.ErrorFillingPaymentInformation
		return nil, err
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
		AmountEuro:     float64(110),
		Name:           PrepareVoucher.VoucherName,
		PhoneNumber:    int64(phoneNumber),
		ProviderId:     providerIdInt,
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
		inTrade[0].Amount = amount.AmountType(storedVoucher.UserPayment.Amount).ToNormalUnit()
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
	body := plutus.ValidateRawTxReq{
		Coin:    voucherData.UserPayment.Coin,
		RawTx:   rawTx,
		Amount:  voucherData.UserPayment.Amount,
		Address: voucherData.UserPayment.Address,
	}
	valid, err := vc.Plutus.ValidateRawTx(body)
	if err != nil || !valid {
		// If fail and coin is POLIS mark as error
		voucherData.Status = hestia.VoucherStatusV2Error
		_, err = vc.Hestia.UpdateVoucherV2(voucherData)
		if err != nil {
			return
		}
		return
	}
	// Broadcast rawTx
	coinConfig, err := coinfactory.GetCoin(voucherData.UserPayment.Coin)
	if err != nil {
		voucherData.Status = hestia.VoucherStatusV2Error
		_, err = vc.Hestia.UpdateVoucherV2(voucherData)
		if err != nil {
			return
		}
		return
	}
	paymentTxid, err, _ := vc.broadCastTxV2(coinConfig, rawTx)
	if err != nil {
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
