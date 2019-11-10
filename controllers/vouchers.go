package controllers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


type VouchersController struct {
	BitcouService    *services.BitcouService
	PreparesVouchers map[string]models.PrepareVoucherInfo
	mapLock sync.RWMutex
}

func (vc *VouchersController) Status(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	status, err := services.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	return status.Vouchers, nil
}

func (vc *VouchersController) GetList(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	vouchersList, err := vc.BitcouService.GetVouchersList()
	if err != nil {
		return nil, err
	}
	return vouchersList, nil
}

func (vc *VouchersController) GetListForPhone(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	vouchersList, err := vc.BitcouService.GetVouchersList()
	if err != nil {
		return nil, err
	}
	vouchersAvailable, err := vc.BitcouService.GetPhoneTopUpList(phoneNb)
	if err != nil {
		return nil, err
	}
	var VouchersList []models.Voucher
	for _, availableVoucher := range vouchersAvailable {
		for _, v := range vouchersList {
			for _, voucher := range v {
				if availableVoucher == voucher.ProductID {
					VouchersList = append(VouchersList, voucher)
				}
			}
		}
	}
	return VouchersList, nil
}

func (vc *VouchersController) Prepare(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	// Get the vouchers percentage fee for PolisPay
	config, err := services.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	// Grab information on the payload
	var PrepareVoucher models.PrepareVoucher
	err = json.Unmarshal(payload, &PrepareVoucher)
	if err != nil {
		return nil, err
	}
	// Create a VoucherID
	newVoucherID := utils.RandomString()
	// Get a payment address from the hot-wallets
	paymentAddr, err := services.GetNewPaymentAddress(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	// If the user is using another coin that is not Polis we will need a Polis payment address to pay the fee
	var feePaymentAddr string
	if PrepareVoucher.Coin != "POLIS" {
		feePaymentAddr, err = services.GetNewPaymentAddress("POLIS")
		if err != nil {
			return nil, err
		}
	}

	// Get the voucher list to grab the name of the voucher.
	vouchers, err := vc.BitcouService.GetVouchersList()
	var selectedVoucher models.Voucher
	for _, voucher := range vouchers[PrepareVoucher.Country] {
		if voucher.ProductID == PrepareVoucher.VoucherType {
			selectedVoucher = voucher
			break
		}
	}
	// Get the paying coin rates
	rates, err := obol.GetCoinRates(obol.ProductionURL, PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}

	// If the user is paying with another coin that is not Polis we will need the Polis rates.
	polisRates, err := obol.GetCoinRates(obol.ProductionURL, "POLIS")
	if err != nil {
		return nil, err
	}
	// Convert the variand id to int
	voucherVariantInt, _ := strconv.Atoi(PrepareVoucher.VoucherVariant)
	// Prepare Tx for Bitcou
	bitcouPrepareTx := models.PurchaseInfo{
		TransactionID: newVoucherID,
		ProductID:     int32(PrepareVoucher.VoucherType),
		VariantID:     int32(voucherVariantInt),
		PhoneNB:       0,
	}
	// Ask bitcou to send amount and address for a specific voucher and add the VoucherID
	purchaseRes, err := vc.BitcouService.GetTransactionInformation(bitcouPrepareTx)
	if err != nil {
		return nil, err
	}
	// Get POLIS EUR rate
	var polisEurRate float64
	for _, rate := range polisRates {
		if rate.Code == "EUR" {
			polisEurRate = rate.Rate
			break
		}
	}
	// Get the coin they are paying to EUR rate
	var paymentCoinEurRates float64
	for _, rate := range rates {
		if rate.Code == "EUR" {
			paymentCoinEurRates = rate.Rate
			break
		}
	}
	amountEuro, _ := strconv.Atoi(purchaseRes.AmountEuro)
	eurAmount := float64(amountEuro / 100)
	// TODO check rates
	// Prepare the payments amount
	var paymentAmount int64
	var feeAmount int64
	if PrepareVoucher.Coin == "POLIS" {
		paymentAmount = int64(((eurAmount/polisEurRate)*1e8)+1) * int64(1+(config.Vouchers.FeePercentage/100))
	} else {
		paymentAmount = int64(((eurAmount / paymentCoinEurRates) * 1e8) + 1)
		feeAmount = int64((eurAmount*float64(config.Vouchers.FeePercentage)/100)/polisEurRate*1e8 + 1)
	}
	paymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: paymentAmount}
	bitcouPaymentInfo := models.PaymentInfo{Address: purchaseRes.Address, Amount: int64(purchaseRes.Amount * 1e8)}
	feeInfo := models.PaymentInfo{Address: feePaymentAddr, Amount: feeAmount}
	// Build the response
	res := models.PrepareVoucherResponse{
		Payment: paymentInfo,
		Fee:     feeInfo,
	}
	// Store on local cache
	prepareVoucher := models.PrepareVoucherInfo{
		ID:             newVoucherID,
		Coin:           PrepareVoucher.Coin,
		Timestamp:      time.Now().Unix(),
		BitcouID:       purchaseRes.BitcouTransactionID,
		VoucherType:    PrepareVoucher.VoucherType,
		VoucherVariant: PrepareVoucher.VoucherVariant,
		Payment:        paymentInfo,
		FeePayment:     feeInfo,
		BitcouPayment:  bitcouPaymentInfo,
		FiatAmount:     int32(eurAmount),
		VoucherName:    selectedVoucher.Name,
	}
	vc.AddVoucherToMap(uid, prepareVoucher)
	return res, nil
}

func (vc *VouchersController) Store(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	var voucherPayments models.StoreVoucher
	err := json.Unmarshal(payload, &voucherPayments)
	if err != nil {
		return nil, err
	}
	storedVoucher, err := vc.GetVoucherFromMap(uid)
	if err != nil {
		return nil, err
	}
	voucher := hestia.Voucher{
		ID:         storedVoucher.ID,
		UID:        uid,
		VoucherID:  storedVoucher.VoucherType,
		VariantID:  storedVoucher.VoucherVariant,
		FiatAmount: storedVoucher.FiatAmount,
		Name:       storedVoucher.VoucherName,
		PaymentData: hestia.Payment{
			Address:       storedVoucher.Payment.Address,
			Amount:        storedVoucher.Payment.Amount,
			Coin:          storedVoucher.Coin,
			RawTx:         voucherPayments.RawTxPayment,
			Txid:          "",
			Confirmations: 0,
		},
		FeePayment: hestia.Payment{
			Address:       storedVoucher.FeePayment.Address,
			Amount:        storedVoucher.FeePayment.Amount,
			Coin:          "polis",
			RawTx:         voucherPayments.RawTxFee,
			Txid:          "",
			Confirmations: 0,
		},
		BitcouPaymentData: hestia.Payment{
			Address:       storedVoucher.BitcouPayment.Address,
			Amount:        storedVoucher.BitcouPayment.Amount,
			Coin:          "dash",
			RawTx:         "",
			Txid:          "",
			Confirmations: 0,
		},
		BitcouID:   storedVoucher.BitcouID,
		RedeemCode: "",
		Status:     hestia.GetVoucherStatusString(hestia.VoucherStatusPending),
		Timestamp:  time.Now().Unix(),
	}
	voucherid, err = services.UpdateVoucher(voucher)
	if err != nil {
		return nil, err
	}
	return voucherid, nil
}

func (vc *VouchersController) Update(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	bearerToken := strings.Split(authToken, "Bearer ")
	if bearerToken[1] != os.Getenv("BITCOU_TOKEN") {
		responses.GlobalResponseNoAuth(c)
		return
	}
	var voucherInfo models.RedeemCodeVoucher
	err := c.BindJSON(&voucherInfo)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	storedVoucherInfo, err := services.GetVoucherInfo(voucherInfo.VoucherID)
	if err != nil {
		responses.GlobalResponseError(nil, errors.New("voucher not found"), c)
		return
	}
	storedVoucherInfo.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusComplete)
	storedVoucherInfo.RedeemCode = voucherInfo.RedeemCode
	storedVoucherInfo.RedeemTimestamp = time.Now().Unix()
	_, err = services.UpdateVoucher(storedVoucherInfo)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	responses.GlobalResponseError("success", nil, c)
	return
}

func (vc *VouchersController) AddVoucherToMap(uid string, voucherPrepare models.PrepareVoucherInfo) {
	vc.mapLock.Lock()
	vc.PreparesVouchers[uid] = voucherPrepare
	vc.mapLock.Unlock()
}

func (vc *VouchersController) RemoveVoucherFromMap(uid string) {
	vc.mapLock.Lock()
	delete(vc.PreparesVouchers, uid)
	vc.mapLock.Unlock()
}

func (vc *VouchersController) GetVoucherFromMap(uid string) (models.PrepareVoucherInfo, error) {
	vc.mapLock.Lock()
	voucher, ok := vc.PreparesVouchers[uid]
	vc.mapLock.Unlock()
	if !ok {
		return models.PrepareVoucherInfo{}, errors.New("voucher not found in cache map")
	}
	return voucher, nil
}
