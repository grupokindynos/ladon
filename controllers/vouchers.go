package controllers

import (
	"github.com/grupokindynos/ladon/services"
)

type VouchersController struct {
	BitcouService *services.BitcouService
}

func (vc *VouchersController) GetVouchersList() (interface{}, error) {
	return vc.BitcouService.GetVouchersList()
}
