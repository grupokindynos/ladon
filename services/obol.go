package services

import (
	"github.com/grupokindynos/common/obol"
)

func GetCoinRates(coin string) ([]obol.Rate, error) {
	return obol.GetCoinRates(coin)
}
