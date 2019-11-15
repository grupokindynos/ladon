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
	"github.com/grupokindynos/olympus-utils/amount"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type VouchersController struct {
	BitcouService    *services.BitcouService
	PreparesVouchers map[string]models.PrepareVoucherInfo
	mapLock          sync.RWMutex
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
	if !config.Vouchers.Available {
		return nil, err
	}
	// Grab information on the payload
	var PrepareVoucher models.PrepareVoucher
	err = json.Unmarshal(payload, &PrepareVoucher)
	if err != nil {
		return nil, err
	}
	coinsConfig, err := services.GetCoinsConfig()
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
	if selectedVoucher.ProductID == 0 {
		return nil, err
	}
	var selectedVariant models.Variants
	for _, variant := range selectedVoucher.Variants {
		if variant.VariantID == PrepareVoucher.VoucherVariant {
			selectedVariant = variant
		}
	}
	if selectedVariant.VariantID == "" {
		return nil, err
	}
	// Convert the variand id to int
	voucherVariantInt, _ := strconv.Atoi(PrepareVoucher.VoucherVariant)
	// Prepare Tx for Bitcou
	if phoneNb == "" {
		phoneNb = "0"
	}
	phoneNumber, err := strconv.Atoi(phoneNb)
	if err != nil {
		return nil, err
	}
	bitcouPrepareTx := models.PurchaseInfo{
		TransactionID: newVoucherID,
		ProductID:     int32(PrepareVoucher.VoucherType),
		VariantID:     int32(voucherVariantInt),
		PhoneNB:       int32(phoneNumber),
	}
	// Ask bitcou to send amount and address for a specific voucher and add the VoucherID
	purchaseRes, err := vc.BitcouService.GetTransactionInformation(bitcouPrepareTx)
	if err != nil {
		return nil, err
	}
	purchaseAmount, err := amount.NewAmount(purchaseRes.Amount)
	if err != nil {
		return nil, err
	}
	// Get the paying coin rates
	var paymentCoinRate float64
	if PrepareVoucher.Coin == "DASH" {
		paymentCoinRate = 1
	} else {
		paymentCoinRate, err = obol.GetCoin2CoinRates(obol.ProductionURL, "DASH", PrepareVoucher.Coin)
		if err != nil {
			return nil, err
		}
	}

	paymentCoinRateAmount, err := amount.NewAmount(paymentCoinRate)
	if err != nil {
		return nil, err
	}
	// Converted amount of the total payed amount on dash to the other crypto.
	// For user usage.
	paymentAmount, err := amount.NewAmount(purchaseAmount.ToNormalUnit() / paymentCoinRateAmount.ToNormalUnit())
	if err != nil {
		return nil, err
	}
	paymentInfo := models.PaymentInfo{Address: paymentAddr, Amount: int64(paymentAmount.ToUnit(amount.AmountSats))}
	// DASH amount on sats to pay Bitcou.
	// For internal usage
	bitcouPaymentInfo := models.PaymentInfo{Address: purchaseRes.Address, Amount: int64(purchaseAmount.ToUnit(amount.AmountSats))}

	var feeInfo, bitcouFeePaymentInfo models.PaymentInfo
	if PrepareVoucher.Coin != "POLIS" {
		// Get the polis rates
		polisRate, err := obol.GetCoin2CoinRates(obol.ProductionURL, "DASH", "POLIS")
		if err != nil {
			return nil, err
		}
		polisRateAmount, err := amount.NewAmount(polisRate)
		if err != nil {
			return nil, err
		}
		feePercentage := float64(paymentCoinConfig.Vouchers.FeePercentage) / float64(100)
		feeAmount, err := amount.NewAmount((purchaseAmount.ToNormalUnit() / polisRateAmount.ToNormalUnit()) * feePercentage)
		if err != nil {
			return nil, err
		}

		// POLIS amount on sats to pay the total fee, this must be at least 4% of the purchased amount for all coins except for Polis.
		// For user usage.
		feeInfo = models.PaymentInfo{Address: feePaymentAddr, Amount: int64(feeAmount.ToUnit(amount.AmountSats))}
		// POLIS amount on sats to pay for the voucher, this must be 4% of the purchased amount for all coins except for Polis.
		// For internal usage
		bitcouFeePercentageOfTotalFee := float64(4) / float64(paymentCoinConfig.Vouchers.FeePercentage)
		bitcouFeePaymentInfo = models.PaymentInfo{Address: "PNTh62FHi2hnSuFwQyjL3ofiVLvrZ9Gzph", Amount: int64(feeAmount.ToUnit(amount.AmountSats) * bitcouFeePercentageOfTotalFee)}
	} else {
		// No Fee if user is paying with POLIS
		feeInfo = models.PaymentInfo{Address: "", Amount: 0}
		bitcouFeePaymentInfo = models.PaymentInfo{Address: "", Amount: 0}
	}

	// Build the response
	res := models.PrepareVoucherResponse{
		Payment: paymentInfo,
		Fee:     feeInfo,
	}
	// Store on local cache
	prepareVoucher := models.PrepareVoucherInfo{
		ID:               newVoucherID,
		Coin:             PrepareVoucher.Coin,
		Timestamp:        time.Now().Unix(),
		BitcouID:         purchaseRes.BitcouTransactionID,
		VoucherType:      PrepareVoucher.VoucherType,
		VoucherVariant:   PrepareVoucher.VoucherVariant,
		Payment:          paymentInfo,
		FeePayment:       feeInfo,
		BitcouPayment:    bitcouPaymentInfo,
		BitcouFeePayment: bitcouFeePaymentInfo,
		VoucherName:      selectedVoucher.Name,
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
		ID:        storedVoucher.ID,
		UID:       uid,
		VoucherID: storedVoucher.VoucherType,
		VariantID: storedVoucher.VoucherVariant,
		Name:      storedVoucher.VoucherName,
		PaymentData: hestia.Payment{
			Address:       storedVoucher.Payment.Address,
			Amount:        storedVoucher.Payment.Amount,
			Coin:          storedVoucher.Coin,
			RawTx:         voucherPayments.RawTx,
			Txid:          "",
			Confirmations: 0,
		},
		FeePayment: hestia.Payment{
			Address:       storedVoucher.FeePayment.Address,
			Amount:        storedVoucher.FeePayment.Amount,
			Coin:          "polis",
			RawTx:         voucherPayments.FeeTx,
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
		BitcouFeePaymentData: hestia.Payment{
			Address:       storedVoucher.BitcouFeePayment.Address,
			Amount:        storedVoucher.BitcouFeePayment.Amount,
			Coin:          "",
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
