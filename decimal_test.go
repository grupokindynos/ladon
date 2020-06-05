package main_test

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/magiconair/properties/assert"
	"github.com/shopspring/decimal"
	"log"
	"testing"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestDecimal(t *testing.T) {
	amount, _ := decimal.NewFromInt(512345678).DivRound(decimal.NewFromInt(1e8), 8).Float64()
	assert.Equal(t, amount, 5.12345678)
}

func TestDecimalToInt(t *testing.T) {
	amount := decimal.NewFromFloat(512345678).DivRound(decimal.NewFromInt(1e8), 8)
	fmt.Println(amount)
	amountSat := amount.Mul(decimal.NewFromInt(1e8))
	fmt.Println(amountSat)
	intAmount := amountSat.IntPart()
	fmt.Println(intAmount)
	assert.Equal(t, intAmount, int64(512345678))
}

func TestDecimalOneLine(t *testing.T) {
	amount := decimal.NewFromFloat(512345678).DivRound(decimal.NewFromInt(1e8), 8)
	fmt.Println(amount)
	amountSat := amount.Mul(decimal.NewFromInt(1e8)).IntPart()
	fmt.Println(amountSat)
	assert.Equal(t, amountSat, int64(512345678))
}
func TestDecimalOneLine2(t *testing.T) {
	amount, _ := decimal.NewFromInt(512345678).DivRound(decimal.NewFromInt(1e8), 8).Float64()
	fmt.Println(amount)
	assert.Equal(t, amount, 5.12345678)
}