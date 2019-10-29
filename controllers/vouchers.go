package controllers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"os"
	"strconv"
	"strings"
	"time"
)

type VouchersController struct {
	BitcouService    *services.BitcouService
	PreparesVouchers map[string]models.PrepareVoucherInfo
}

func (vc *VouchersController) GetServiceStatus(payload []byte, uid string, voucherid string) (interface{}, error) {
	status, err := services.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	return jwt.EncryptJWE(uid, status.Vouchers)
}

func (vc *VouchersController) GetList(payload []byte, uid string, voucherid string) (interface{}, error) {
	vouchersList, err := vc.BitcouService.GetVouchersList()
	if err != nil {
		return nil, err
	}
	return jwt.EncryptJWE(uid, vouchersList)
}

func (vc *VouchersController) GetToken(payload []byte, uid string, voucherid string) (interface{}, error) {
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
	// Get the Dash rates to calculate Bitcou payment
	dashRates, err := obol.GetCoinRates(obol.ProductionURL, "DASH")
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
	// Get DASH EUR rate
	var dashEurRate float64
	for _, rate := range dashRates {
		if rate.Code == "EUR" {
			dashEurRate = rate.Rate
			break
		}
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
	// Sanitize Bitcou float response
	voucherPriceSats := int32(purchaseRes.Amount*1e8) + 1
	voucherPriceTrunk := float64(voucherPriceSats) / 1e8

	// Get the price of the voucher on EUR
	voucherEurPrice := float64(int((dashEurRate*voucherPriceTrunk)*100)) / 100

	// Prepare the payments amount
	var paymentAmount int64
	var feeAmount int64
	if PrepareVoucher.Coin == "POLIS" {
		paymentAmount = int64(((voucherEurPrice/polisEurRate)*1e8)+1) * int64(1+(config.Vouchers.FeePercentage/100))
	} else {
		paymentAmount = int64(((voucherEurPrice / paymentCoinEurRates) * 1e8) + 1)
		feeAmount = int64((voucherEurPrice*float64(config.Vouchers.FeePercentage)/100)/polisEurRate*1e8 + 1)
	}
	paymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: paymentAmount}
	bitcouPaymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: paymentAmount}
	feeInfo := models.PaymentInfo{Address: feePaymentAddr, Amount: feeAmount}
	// Build the response
	res := models.PrepareVoucherResponse{
		Payment: paymentInfo,
		Fee:     feeInfo,
	}
	// Store on local cache
	vc.PreparesVouchers[uid] = models.PrepareVoucherInfo{
		ID:             newVoucherID,
		Coin:           PrepareVoucher.Coin,
		Timestamp:      time.Now().Unix(),
		VoucherType:    PrepareVoucher.VoucherType,
		VoucherVariant: PrepareVoucher.VoucherVariant,
		Payment:        paymentInfo,
		FeePayment:     feeInfo,
		BitcouPayment:  bitcouPaymentInfo,
		FiatAmount:     int32(voucherEurPrice),
		VoucherName:    selectedVoucher.Name,
	}
	return jwt.EncryptJWE(uid, res)
}

func (vc *VouchersController) Store(payload []byte, uid string, voucherid string) (interface{}, error) {
	var voucherPayments models.StoreVoucher
	err := json.Unmarshal(payload, &voucherPayments)
	if err != nil {
		return nil, err
	}
	storedVoucher := vc.PreparesVouchers[uid]
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
		RedeemCode: "",
		Status:     hestia.GetVoucherStatusString(hestia.VoucherStatusPending),
		Timestamp:  time.Now().Unix(),
	}
	voucherid, err = services.UpdateVoucher(voucher)
	if err != nil {
		return nil, err
	}
	return jwt.EncryptJWE(uid, voucherid)
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
