// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/grupokindynos/ladon/services (interfaces: HestiaService,PlutusService,BitcouService)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	hestia "github.com/grupokindynos/common/hestia"
	plutus "github.com/grupokindynos/common/plutus"
	models "github.com/grupokindynos/ladon/models"
	reflect "reflect"
)

// MockHestiaService is a mock of HestiaService interface
type MockHestiaService struct {
	ctrl     *gomock.Controller
	recorder *MockHestiaServiceMockRecorder
}

// MockHestiaServiceMockRecorder is the mock recorder for MockHestiaService
type MockHestiaServiceMockRecorder struct {
	mock *MockHestiaService
}

// NewMockHestiaService creates a new mock instance
func NewMockHestiaService(ctrl *gomock.Controller) *MockHestiaService {
	mock := &MockHestiaService{ctrl: ctrl}
	mock.recorder = &MockHestiaServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHestiaService) EXPECT() *MockHestiaServiceMockRecorder {
	return m.recorder
}

// GetCoinsConfig mocks base method
func (m *MockHestiaService) GetCoinsConfig() ([]hestia.Coin, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCoinsConfig")
	ret0, _ := ret[0].([]hestia.Coin)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCoinsConfig indicates an expected call of GetCoinsConfig
func (mr *MockHestiaServiceMockRecorder) GetCoinsConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCoinsConfig", reflect.TypeOf((*MockHestiaService)(nil).GetCoinsConfig))
}

// GetVoucherInfo mocks base method
func (m *MockHestiaService) GetVoucherInfo(arg0 string) (hestia.Voucher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVoucherInfo", arg0)
	ret0, _ := ret[0].(hestia.Voucher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVoucherInfo indicates an expected call of GetVoucherInfo
func (mr *MockHestiaServiceMockRecorder) GetVoucherInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVoucherInfo", reflect.TypeOf((*MockHestiaService)(nil).GetVoucherInfo), arg0)
}

// GetVouchersByTimestamp mocks base method
func (m *MockHestiaService) GetVouchersByTimestamp(arg0, arg1 string) ([]hestia.Voucher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVouchersByTimestamp", arg0, arg1)
	ret0, _ := ret[0].([]hestia.Voucher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVouchersByTimestamp indicates an expected call of GetVouchersByTimestamp
func (mr *MockHestiaServiceMockRecorder) GetVouchersByTimestamp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVouchersByTimestamp", reflect.TypeOf((*MockHestiaService)(nil).GetVouchersByTimestamp), arg0, arg1)
}

// GetVouchersStatus mocks base method
func (m *MockHestiaService) GetVouchersStatus() (hestia.Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVouchersStatus")
	ret0, _ := ret[0].(hestia.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVouchersStatus indicates an expected call of GetVouchersStatus
func (mr *MockHestiaServiceMockRecorder) GetVouchersStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVouchersStatus", reflect.TypeOf((*MockHestiaService)(nil).GetVouchersStatus))
}

// UpdateVoucher mocks base method
func (m *MockHestiaService) UpdateVoucher(arg0 hestia.Voucher) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateVoucher", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateVoucher indicates an expected call of UpdateVoucher
func (mr *MockHestiaServiceMockRecorder) UpdateVoucher(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateVoucher", reflect.TypeOf((*MockHestiaService)(nil).UpdateVoucher), arg0)
}

// MockPlutusService is a mock of PlutusService interface
type MockPlutusService struct {
	ctrl     *gomock.Controller
	recorder *MockPlutusServiceMockRecorder
}

// MockPlutusServiceMockRecorder is the mock recorder for MockPlutusService
type MockPlutusServiceMockRecorder struct {
	mock *MockPlutusService
}

// NewMockPlutusService creates a new mock instance
func NewMockPlutusService(ctrl *gomock.Controller) *MockPlutusService {
	mock := &MockPlutusService{ctrl: ctrl}
	mock.recorder = &MockPlutusServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlutusService) EXPECT() *MockPlutusServiceMockRecorder {
	return m.recorder
}

// GetNewPaymentAddress mocks base method
func (m *MockPlutusService) GetNewPaymentAddress(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNewPaymentAddress", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNewPaymentAddress indicates an expected call of GetNewPaymentAddress
func (mr *MockPlutusServiceMockRecorder) GetNewPaymentAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNewPaymentAddress", reflect.TypeOf((*MockPlutusService)(nil).GetNewPaymentAddress), arg0)
}

// GetWalletBalance mocks base method
func (m *MockPlutusService) GetWalletBalance(arg0 string) (plutus.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWalletBalance", arg0)
	ret0, _ := ret[0].(plutus.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWalletBalance indicates an expected call of GetWalletBalance
func (mr *MockPlutusServiceMockRecorder) GetWalletBalance(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWalletBalance", reflect.TypeOf((*MockPlutusService)(nil).GetWalletBalance), arg0)
}

// SubmitPayment mocks base method
func (m *MockPlutusService) SubmitPayment(arg0 plutus.SendAddressBodyReq) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitPayment", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubmitPayment indicates an expected call of SubmitPayment
func (mr *MockPlutusServiceMockRecorder) SubmitPayment(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitPayment", reflect.TypeOf((*MockPlutusService)(nil).SubmitPayment), arg0)
}

// ValidateRawTx mocks base method
func (m *MockPlutusService) ValidateRawTx(arg0 plutus.ValidateRawTxReq) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateRawTx", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateRawTx indicates an expected call of ValidateRawTx
func (mr *MockPlutusServiceMockRecorder) ValidateRawTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateRawTx", reflect.TypeOf((*MockPlutusService)(nil).ValidateRawTx), arg0)
}

// MockBitcouService is a mock of BitcouService interface
type MockBitcouService struct {
	ctrl     *gomock.Controller
	recorder *MockBitcouServiceMockRecorder
}

// MockBitcouServiceMockRecorder is the mock recorder for MockBitcouService
type MockBitcouServiceMockRecorder struct {
	mock *MockBitcouService
}

// NewMockBitcouService creates a new mock instance
func NewMockBitcouService(ctrl *gomock.Controller) *MockBitcouService {
	mock := &MockBitcouService{ctrl: ctrl}
	mock.recorder = &MockBitcouServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBitcouService) EXPECT() *MockBitcouServiceMockRecorder {
	return m.recorder
}

// GetPhoneTopUpList mocks base method
func (m *MockBitcouService) GetPhoneTopUpList(arg0 string) ([]int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPhoneTopUpList", arg0)
	ret0, _ := ret[0].([]int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPhoneTopUpList indicates an expected call of GetPhoneTopUpList
func (mr *MockBitcouServiceMockRecorder) GetPhoneTopUpList(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPhoneTopUpList", reflect.TypeOf((*MockBitcouService)(nil).GetPhoneTopUpList), arg0)
}

// GetTransactionInformation mocks base method
func (m *MockBitcouService) GetTransactionInformation(arg0 models.PurchaseInfo) (models.PurchaseInfoResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTransactionInformation", arg0)
	ret0, _ := ret[0].(models.PurchaseInfoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTransactionInformation indicates an expected call of GetTransactionInformation
func (mr *MockBitcouServiceMockRecorder) GetTransactionInformation(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransactionInformation", reflect.TypeOf((*MockBitcouService)(nil).GetTransactionInformation), arg0)
}
