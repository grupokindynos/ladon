package controllers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/responses"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"os"
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
	var PrepareVoucher models.PrepareVoucher
	err := json.Unmarshal(payload, &PrepareVoucher)
	if err != nil {
		return nil, err
	}
	addr, err := services.GetNewPaymentAddress(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	vouchers, err := vc.BitcouService.GetVouchersList()
	var selectedVoucher models.Voucher
	for _, voucher := range vouchers[PrepareVoucher.Country] {
		if voucher.ProductID == PrepareVoucher.VoucherType {
			selectedVoucher = voucher
			break
		}
	}
	var voucherVariant models.Variants
	for _, variant := range selectedVoucher.Variants {
		if variant.VariantID == PrepareVoucher.VoucherVariant {
			voucherVariant = variant
			break
		}
	}
	rates, err := obol.GetCoinRates(PrepareVoucher.Coin)
	if err != nil {
		return nil, err
	}
	var eurRate float64
	for _, rate := range rates {
		if rate.Code == "EUR" {
			eurRate = rate.Rate
			break
		}
	}
	voucherPrice := voucherVariant.Price / 100
	amount := int32(( float64(voucherPrice) / eurRate ) * 1e8 + 1)
	res := models.PrepareVoucherResponse{
		Address: addr,
		Amount:  amount,
	}
	vc.PreparesVouchers[uid] = models.PrepareVoucherInfo{
		Coin:           PrepareVoucher.Coin,
		Timestamp:      time.Now().Unix(),
		VoucherType:    PrepareVoucher.VoucherType,
		VoucherVariant: PrepareVoucher.VoucherVariant,
		Address:        res.Address,
		Amount:         res.Amount,
		FiatAmount:     voucherVariant.Price,
		VoucherName:    selectedVoucher.Name,
	}
	return jwt.EncryptJWE(uid, res)
}

func (vc *VouchersController) Store(payload []byte, uid string, voucherid string) (interface{}, error) {
	var rawTx string
	err := json.Unmarshal(payload, &rawTx)
	if err != nil {
		return nil, err
	}
	storedVoucher := vc.PreparesVouchers[uid]
	id := sha256.Sum256([]byte(rawTx))
	voucher := hestia.Voucher{
		ID:         string(id[:]),
		UID:        uid,
		VoucherID:  storedVoucher.VoucherType,
		VariantID:  storedVoucher.VoucherVariant,
		FiatAmount: storedVoucher.FiatAmount,
		Name:       storedVoucher.VoucherName,
		PaymentData: hestia.Payment{
			Address:       storedVoucher.Address,
			Amount:        storedVoucher.Amount,
			Coin:          storedVoucher.Coin,
			RawTx:         rawTx,
			Txid:          "",
			Confirmations: 0,
		},
		BitcouPaymentData: hestia.Payment{},
		RedeemCode:        "",
		Status:            "PENDING",
		Timestamp:         string(time.Now().Unix()),
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
	storedVoucherInfo.Status = "COMPLETE"
	storedVoucherInfo.RedeemCode = voucherInfo.RedeemCode
	_, err = services.UpdateVoucher(storedVoucherInfo)
	if err != nil {
		responses.GlobalResponseError(nil, err, c)
		return
	}
	responses.GlobalResponseError("success", nil, c)
	return
}
