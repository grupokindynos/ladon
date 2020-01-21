package test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	commonErrors "github.com/grupokindynos/common/errors"
	"github.com/grupokindynos/common/hestia"
	obolMocks "github.com/grupokindynos/common/obol/mocks"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/ladon/controllers"
	"github.com/grupokindynos/ladon/mocks"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ogen-utils/amount"
)

func TestStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	emptyHestiaConfig := hestia.Config{}
	vouchersAvailable := hestia.Config{Vouchers: hestia.Available{Service: true}}
	testError := errors.New("testing error")
	payload := []byte{10, 10, 10}
	voucherid := "123voucher"
	uid := "123userid"
	phoneNb := "123phone"

	mockHestiaService := mocks.NewMockHestiaService(mockCtrl)
	vouchersCtrl := &controllers.VouchersController{Hestia: mockHestiaService}

	gomock.InOrder(
		mockHestiaService.EXPECT().GetVouchersStatus().Return(vouchersAvailable, nil),
		mockHestiaService.EXPECT().GetVouchersStatus().Return(emptyHestiaConfig, testError),
	)

	available, err := vouchersCtrl.Status(payload, uid, voucherid, phoneNb)

	// Test vouchers available
	if err != nil {
		t.Fatal("Test vouchers available - error returned not equal to nil")
	}

	if available == false {
		t.Fatal("Test vouchers available - vouchers returned not available")
	}

	available, err = vouchersCtrl.Status(payload, uid, voucherid, phoneNb)

	// Test vouchers not available
	if err != testError {
		t.Fatal("Test vouchers not available - mismatching returned error")
	}

	if available != nil {
		t.Fatal("Test vouchers not available - avaialability not equal to nil")
	}
}

func TestPrepare(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// common
	prepareVouchersMap := make(map[string]models.PrepareVoucherInfo)
	vouchersAvailable := hestia.Config{Vouchers: hestia.Available{Service: true}}
	paymentAddress := "123payhere456"
	feePaymentAddress := "123payherethefee456"
	voucherid := "123voucher"
	uid := "123userid"
	phoneNb := "12345609"
	enoughBalance := plutus.Balance{
		Confirmed: 15.6,
	}
	notEnoughBalance := plutus.Balance{
		Confirmed: 5.55,
	}
	voucherProp := hestia.Properties{
		FeePercentage: 50,
		Available:     true,
	}
	coinsConfig := []hestia.Coin{
		hestia.Coin{Ticker: "BTC", Vouchers: voucherProp},
		hestia.Coin{Ticker: "POLIS", Vouchers: voucherProp},
		hestia.Coin{Ticker: "XSG", Vouchers: voucherProp},
		hestia.Coin{Ticker: "DASH", Vouchers: voucherProp},
	}
	purchaseInfo := models.PurchaseInfoResponse{
		Amount:              10.56,
		AmountEuro:          "1.5",
		BitcouTransactionID: "123456",
		Address:             "123payheretobitcou456",
	}
	purchaseAmount, _ := amount.NewAmount(purchaseInfo.Amount)

	// Dash
	prepareVoucherDash := models.PrepareVoucher{
		Coin:           "DASH",
		VoucherVariant: "123456",
	}
	payloadDash, _ := json.Marshal(prepareVoucherDash)
	dashPolisRate := 1.54
	paymentAmountDash := purchaseAmount
	dashPolisRateAmount, _ := amount.NewAmount(dashPolisRate)
	feePercentageDash := float64(coinsConfig[3].Vouchers.FeePercentage) / float64(100)
	feeAmountDash, _ := amount.NewAmount((purchaseAmount.ToNormalUnit() / dashPolisRateAmount.ToNormalUnit()) * feePercentageDash)
	paymentInfoDash := models.PaymentInfo{
		Address: paymentAddress,
		Amount:  int64(paymentAmountDash.ToUnit(amount.AmountSats)),
	}
	feePaymentInfoDash := models.PaymentInfo{
		Address: feePaymentAddress,
		Amount:  int64(feeAmountDash.ToUnit(amount.AmountSats)),
	}
	vouchersResponseDash := models.PrepareVoucherResponse{
		Payment: paymentInfoDash,
		Fee:     feePaymentInfoDash,
	}

	// BTC
	prepareVoucherBTC := models.PrepareVoucher{
		Coin:           "BTC",
		VoucherVariant: "123456",
	}
	payloadBTC, _ := json.Marshal(prepareVoucherBTC)
	dashBTCRate := 15.45

	mockHestiaService := mocks.NewMockHestiaService(mockCtrl)
	mockPlutusService := mocks.NewMockPlutusService(mockCtrl)
	mockBitcouService := mocks.NewMockBitcouService(mockCtrl)
	mockObolService := obolMocks.NewMockObolService(mockCtrl)

	vouchersCtrl := &controllers.VouchersController{
		PreparesVouchers: prepareVouchersMap,
		Plutus:           mockPlutusService,
		Hestia:           mockHestiaService,
		Bitcou:           mockBitcouService,
		Obol:             mockObolService,
	}

	gomock.InOrder(
		// calls test returned response with enough balance
		mockHestiaService.EXPECT().GetVouchersStatus().Return(vouchersAvailable, nil),
		mockHestiaService.EXPECT().GetCoinsConfig().Return(coinsConfig, nil),
		mockPlutusService.EXPECT().GetNewPaymentAddress(gomock.Eq("DASH")).Return(paymentAddress, nil),
		mockPlutusService.EXPECT().GetNewPaymentAddress(gomock.Eq("POLIS")).Return(feePaymentAddress, nil),
		mockBitcouService.EXPECT().GetTransactionInformation(gomock.Any()).Return(purchaseInfo, nil),
		mockPlutusService.EXPECT().GetWalletBalance(gomock.Eq("DASH")).Return(enoughBalance, nil),
		mockObolService.EXPECT().GetCoin2CoinRates(gomock.Eq("DASH"), "POLIS").Return(dashPolisRate, nil),
		mockHestiaService.EXPECT().GetVouchersByTimestamp(gomock.Any(), gomock.Any()).Return([]hestia.Voucher{},  nil),

		// calls test returned response with not enough balance
		mockHestiaService.EXPECT().GetVouchersStatus().Return(vouchersAvailable, nil),
		mockHestiaService.EXPECT().GetCoinsConfig().Return(coinsConfig, nil),
		mockPlutusService.EXPECT().GetNewPaymentAddress(gomock.Eq("BTC")).Return(paymentAddress, nil),
		mockPlutusService.EXPECT().GetNewPaymentAddress(gomock.Eq("POLIS")).Return(feePaymentAddress, nil),
		mockBitcouService.EXPECT().GetTransactionInformation(gomock.Any()).Return(purchaseInfo, nil),
		mockObolService.EXPECT().GetCoin2CoinRates(gomock.Eq("DASH"), gomock.Eq("BTC")).Return(dashBTCRate, nil),
		mockPlutusService.EXPECT().GetWalletBalance(gomock.Eq("DASH")).Return(notEnoughBalance, nil),

	)

	// test returned response with enough balance
	res, err := vouchersCtrl.Prepare(payloadDash, uid, voucherid, phoneNb)

	if err != nil {
		t.Fatal("test returned response - err not equal to nil - " + err.Error())
	}

	if res != vouchersResponseDash {
		t.Fatal("test returned response - mismatching voucher response")
	}

	// test returned response with not enough balance
	res, err = vouchersCtrl.Prepare(payloadBTC, uid, voucherid, phoneNb)

	if err != commonErrors.ErrorNotEnoughDash {
		t.Fatal("test returned response with not enough balance - mismatching returned error")
	}

	if res != nil {
		t.Fatal("test returned response with not enough balance - response is not equal to nil")
	}
}
