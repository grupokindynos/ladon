package controllers

import (
	"encoding/json"
	"errors"
	"github.com/grupokindynos/common/blockbook"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	commonErrors "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/common/utils"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"github.com/olympus-protocol/ogen/utils/amount"
)

type VouchersController struct {
	PreparesVouchers map[string]models.PrepareVoucherInfo
	mapLock          sync.RWMutex
	TxsAvailable     bool
	Plutus           services.PlutusService
	Hestia           services.HestiaService
	Bitcou           services.BitcouService
	Obol             obol.ObolService
}

func (vc *VouchersController) Status(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	status, err := vc.Hestia.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	return status.Vouchers.Service, nil
}

func (vc *VouchersController) GetListForPhone(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	vouchersAvailable, err := vc.Bitcou.GetPhoneTopUpList(phoneNb)
	if err != nil {
		return nil, err
	}
	return vouchersAvailable, nil
}

func (vc *VouchersController) Prepare(payload []byte, uid string, voucherid string, phoneNb string) (interface{}, error) {
	// Get the vouchers percentage fee for PolisPay
	config, err := vc.Hestia.GetVouchersStatus()
	if err != nil {
		return nil, err
	}
	if !config.Vouchers.Service {
		return nil, err
	}
	// Grab information on the payload
	var PrepareVoucher models.PrepareVoucher
	err = json.Unmarshal(payload, &PrepareVoucher)
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
	// If the user is using another coin that is not Polis we will need a Polis payment address to pay the fee
	var feePaymentAddr string
	if PrepareVoucher.Coin != "POLIS" {
		feePaymentAddr, err = vc.Plutus.GetNewPaymentAddress("POLIS")
		if err != nil {
			return nil, err
		}
	}

	// Convert the variant id to int
	voucherVariantInt, _ := strconv.Atoi(PrepareVoucher.VoucherVariant)
	// Prepare Tx for Bitcou
	if PrepareVoucher.PhoneNumber == "" {
		PrepareVoucher.PhoneNumber = "0"
	}
	phoneNumber, err := strconv.Atoi(PrepareVoucher.PhoneNumber)
	if err != nil {
		return nil, err
	}
	bitcouPrepareTx := models.PurchaseInfo{
		TransactionID: newVoucherID,
		ProductID:     int32(PrepareVoucher.VoucherType),
		VariantID:     int32(voucherVariantInt),
		PhoneNB:       int64(phoneNumber),
	}

	// Ask bitcou to send amount and address for a specific voucher and add the VoucherID
	purchaseRes, err := vc.Bitcou.GetTransactionInformation(bitcouPrepareTx)
	if err != nil {
		return nil, err
	}

	purchaseAmount, err := amount.NewAmount(purchaseRes.Amount)
	if err != nil {
		return nil, err
	}

	purchaseAmountEuro, err := strconv.ParseFloat(purchaseRes.AmountEuro, 64)
	if err != nil {
		return nil, err
	}

	purchaseAmountEuro = purchaseAmountEuro / 100
	euroRate := purchaseAmountEuro / purchaseAmount.ToNormalUnit()

	// Get the paying coin rates
	var paymentCoinRate float64
	if PrepareVoucher.Coin == "DASH" {
		paymentCoinRate = 1
	} else {
		paymentCoinRate, err = vc.Obol.GetCoin2CoinRates(PrepareVoucher.Coin, "DASH")
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

	// Get our current dash balance
	balance, err := vc.Plutus.GetWalletBalance("DASH")
	if err != nil {
		return nil, err
	}

	dashAmount, err := amount.NewAmount(balance.Confirmed)
	if err != nil {
		return nil, err
	}

	dashBalance := int64(dashAmount.ToUnit(amount.AmountSats))

	// Check if we have enough dash to pay the voucher to bitcou
	if PrepareVoucher.Coin != "DASH" && dashBalance < bitcouPaymentInfo.Amount {
		return nil, commonErrors.ErrorNotEnoughDash
	}

	var feeInfo, bitcouFeePaymentInfo models.PaymentInfo
	var feeAmountEuro float64

	if PrepareVoucher.Coin != "POLIS" {
		// Get the polis rates
		polisRate, err := vc.Obol.GetCoin2CoinRates("POLIS", "DASH")
		if err != nil {
			return nil, err
		}
		polisRateAmount, err := amount.NewAmount(polisRate)
		if err != nil {
			return nil, err
		}
		feePercentage := paymentCoinConfig.Vouchers.FeePercentage / float64(100)
		feeAmount, err := amount.NewAmount((purchaseAmount.ToNormalUnit() / polisRateAmount.ToNormalUnit()) * feePercentage) // purchaseAmount in DASH gets converted to Polis. This is the fee payment in Polis.
		if err != nil {
			return nil, err
		}
		feeAmountEuro = feeAmount.ToNormalUnit() * polisRateAmount.ToNormalUnit() * euroRate

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
		feeAmountEuro = 0
		bitcouFeePaymentInfo = models.PaymentInfo{Address: "", Amount: 0}
	}

	// Validate that users hasn't bought more than 210 euro in vouchers on the last 24 hours.
	timestamp := strconv.FormatInt(time.Now().Unix()-24*3600, 10)

	vouchers, err := vc.Hestia.GetVouchersByTimestamp(uid, timestamp)
	if err != nil {
		return nil, err
	}

	totalAmountEuro := purchaseAmountEuro

	for _, voucher := range vouchers {
		if voucher.Status != hestia.GetVoucherStatusString(hestia.VoucherStatusError) && voucher.Status != hestia.GetVoucherStatusString(hestia.VoucherStatusRefunded) { // Excludes errored Vouchers from Spent Amount
			amEr, _ := strconv.ParseFloat(voucher.AmountEuro, 64)
			amEr /= 100
			//amFeeEr, _ := strconv.ParseFloat(voucher.AmountFeeEuro, 64)
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
		AmountEuro:       purchaseRes.AmountEuro,
		AmountFeeEuro:    strconv.FormatFloat(feeAmountEuro, 'f', 6, 64),
		Name:             PrepareVoucher.VoucherName,
		Image:            PrepareVoucher.VoucherImage,
		PhoneNumber:      int64(phoneNumber),
		ProviderId:       PrepareVoucher.ProviderId,
		Valid:            PrepareVoucher.Valid,
	}

	vc.AddVoucherToMap(uid, prepareVoucher)
	return res, nil
}

func (vc *VouchersController) Store(payload []byte, uid string, voucherId string, phoneNb string) (interface{}, error) {
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
		PaymentData: hestia.Payment{
			Address:       storedVoucher.Payment.Address,
			Amount:        storedVoucher.Payment.Amount,
			Coin:          storedVoucher.Coin,
			Txid:          "",
			Confirmations: 0,
		},
		FeePayment: hestia.Payment{
			Address:       storedVoucher.FeePayment.Address,
			Amount:        storedVoucher.FeePayment.Amount,
			Coin:          "polis",
			Txid:          "",
			Confirmations: 0,
		},
		BitcouPaymentData: hestia.Payment{
			Address:       storedVoucher.BitcouPayment.Address,
			Amount:        storedVoucher.BitcouPayment.Amount,
			Coin:          "dash",
			Txid:          "",
			Confirmations: 0,
		},
		BitcouFeePaymentData: hestia.Payment{
			Address:       storedVoucher.BitcouFeePayment.Address,
			Amount:        storedVoucher.BitcouFeePayment.Amount,
			Coin:          "",
			Txid:          "",
			Confirmations: 0,
		},
		BitcouID:      storedVoucher.BitcouID,
		RedeemCode:    "",
		RefundAddr:    voucherPayments.RefundAddr,
		RefundFeeAddr: voucherPayments.RefundFeeAddr,
		Status:        hestia.GetVoucherStatusString(hestia.VoucherStatusPending),
		Timestamp:     time.Now().Unix(),
		AmountEuro:    storedVoucher.AmountEuro,
		AmountFeeEuro: storedVoucher.AmountFeeEuro,
		Name:          storedVoucher.Name,
		Image:         storedVoucher.Image,
		PhoneNumber:   storedVoucher.PhoneNumber,
		ProviderId:    storedVoucher.ProviderId,
		Valid:         storedVoucher.Valid,
	}

	vc.RemoveVoucherFromMap(uid)
	voucherId, err = vc.Hestia.UpdateVoucher(voucher)
	if err != nil {
		return nil, err
	}
	go vc.decodeAndCheckTx(voucher, storedVoucher, voucherPayments.RawTx, voucherPayments.FeeTx)
	return voucherId, nil
}

func (vc *VouchersController) decodeAndCheckTx(voucherData hestia.Voucher, storedVoucherData models.PrepareVoucherInfo, rawTx string, feeTx string) {

	var FeeTxId string
	if voucherData.PaymentData.Coin != "POLIS" {
		// Validate feeRawTx
		body := plutus.ValidateRawTxReq{
			Coin:    voucherData.FeePayment.Coin,
			RawTx:   feeTx,
			Amount:  voucherData.FeePayment.Amount,
			Address: voucherData.FeePayment.Address,
		}
		valid, err := vc.Plutus.ValidateRawTx(body)
		if err != nil || !valid {
			// If fail, we should mark error, no spent anything.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
			_, err = vc.Hestia.UpdateVoucher(voucherData)
			if err != nil {
				return
			}
			return
		}
		// Broadcast fee rawTx
		polisCoinConfig, err := coinfactory.GetCoin("POLIS")
		if err != nil {
			// If get coin fail, we should mark error, no spent anything.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
			_, err = vc.Hestia.UpdateVoucher(voucherData)
			if err != nil {
				return
			}
			return
		}
		FeeTxId, err = vc.broadCastTx(polisCoinConfig, feeTx)
		if err != nil {
			// If broadcast fail, we should mark error, no spent anything.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
			_, err = vc.Hestia.UpdateVoucher(voucherData)
			if err != nil {
				return
			}
			return
		}
	}
	// Validate Payment RawTx
	body := plutus.ValidateRawTxReq{
		Coin:    voucherData.PaymentData.Coin,
		RawTx:   rawTx,
		Amount:  voucherData.PaymentData.Amount,
		Address: voucherData.PaymentData.Address,
	}
	valid, err := vc.Plutus.ValidateRawTx(body)
	if err != nil || !valid {
		// If fail and coin is POLIS mark as error
		voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
		if voucherData.PaymentData.Coin != "POLIS" {
			// If decode fail and coin is different than POLIS we should mark refund fees.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundFee)
		}
		_, err = vc.Hestia.UpdateVoucher(voucherData)
		if err != nil {
			return
		}
		return
	}
	// Broadcast rawTx
	coinConfig, err := coinfactory.GetCoin(voucherData.PaymentData.Coin)
	if err != nil {
		// If get coin fail and coin is POLIS mark as error
		voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
		if voucherData.PaymentData.Coin != "POLIS" {
			// If get coin fail and coin is different than POLIS we should mark refund fees.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundFee)
		}
		_, err = vc.Hestia.UpdateVoucher(voucherData)
		if err != nil {
			return
		}
		return
	}
	paymentTxid, err := vc.broadCastTx(coinConfig, rawTx)
	if err != nil {
		// If broadcast fail and coin is POLIS mark as error
		voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
		if voucherData.PaymentData.Coin != "POLIS" {
			// If broadcast fail and coin is different than POLIS we should mark refund fees.
			voucherData.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundFee)
		}
		_, err = vc.Hestia.UpdateVoucher(voucherData)
		if err != nil {
			return
		}
		return
	}
	// Update voucher model include txid.
	voucherData.PaymentData.Txid = paymentTxid
	voucherData.FeePayment.Txid = FeeTxId
	_, err = vc.Hestia.UpdateVoucher(voucherData)
	if err != nil {
		return
	}
}

func (vc *VouchersController) broadCastTx(coinConfig *coins.Coin, rawTx string) (txid string, err error) {
	if !vc.TxsAvailable {
		return "not published due no-txs flag", nil
	}
	blockbookWrapper := blockbook.NewBlockBookWrapper(coinConfig.Info.Blockbook)
	return blockbookWrapper.SendTx(rawTx)
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
	storedVoucherInfo, err := vc.Hestia.GetVoucherInfo(voucherInfo.VoucherID)
	if err != nil {
		responses.GlobalResponseError(nil, errors.New("voucher not found"), c)
		return
	}
	storedVoucherInfo.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusComplete)
	storedVoucherInfo.RedeemCode = voucherInfo.RedeemCode
	storedVoucherInfo.RedeemTimestamp = time.Now().Unix()
	_, err = vc.Hestia.UpdateVoucher(storedVoucherInfo)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	// Submit Bitcou Fee
	if storedVoucherInfo.PaymentData.Coin != "POLIS" {
		amountHand := amount.AmountType(storedVoucherInfo.BitcouFeePaymentData.Amount)
		bitcouPayment := plutus.SendAddressBodyReq{
			Address: storedVoucherInfo.BitcouFeePaymentData.Address,
			Coin:    "POLIS",
			Amount:  amountHand.ToNormalUnit(),
		}
		// TODO make sure error is handled
		txid, _ := vc.Plutus.SubmitPayment(bitcouPayment)
		storedVoucherInfo.BitcouFeePaymentData.Txid = txid
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

func (vc *VouchersController) GetUserVouchersByTimestampOld(uid string, timestamp string) (vouchers []hestia.Voucher, err error) {
	req, err := mvt.CreateMVTToken("GET", hestia.ProductionURL+"/voucher/all_by_timestamp?timestamp="+timestamp+"&userid="+uid, "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: 40 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return nil, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return nil, errors.New("no header signature")
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return nil, err
	}
	var response []hestia.Voucher
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
