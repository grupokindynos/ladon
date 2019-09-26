package controllers

import (
	"errors"
	"github.com/grupokindynos/common/jwt"
	"github.com/grupokindynos/ladon/services"
)

type VouchersController struct {
	BitcouService *services.BitcouService
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

func (vc *VouchersController) GetInfo(payload []byte, uid string, voucherid string) (interface{}, error) {
	if voucherid == "" {
		return nil, errors.New("no voucher id provided")
	}
	voucher, err := services.GetVoucherInfo(voucherid)
	if err != nil {
		return nil, err
	}
	return jwt.EncryptJWE(uid, voucher)
}

func (vc *VouchersController) GetToken(payload []byte, uid string, voucherid string) (interface{}, error) {
	return jwt.EncryptJWE(uid, nil)
}

func (vc *VouchersController) Store(payload []byte, uid string, voucherid string) (interface{}, error) {
	return jwt.EncryptJWE(uid, nil)
}

func (vc *VouchersController) Update(payload []byte, uid string, voucherid string) (interface{}, error) {
	return jwt.EncryptJWE(uid, nil)
}
