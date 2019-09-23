package controllers

import (
	"github.com/grupokindynos/ladon/services"
)

type VouchersController struct {
	BitcouService *services.BitcouService
}

func (vc *VouchersController) GetServiceStatus() (interface{}, error) {
	return nil, nil
}

func (vc *VouchersController) GetList() (interface{}, error) {
	return vc.BitcouService.GetVouchersList()
}

func (vc *VouchersController) GetInfo() (interface{}, error) {
	return nil, nil
}

func (vc *VouchersController) GetToken() (interface{}, error) {
	return nil, nil
}

func (vc *VouchersController) Store() (interface{}, error) {
	return nil, nil
}

func (vc *VouchersController) Update() (interface{}, error) {
	return nil, nil
}
