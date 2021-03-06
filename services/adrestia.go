package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
)

type AdrestiaRequests struct {
	AdrestiaUrl string
}

func (a *AdrestiaRequests) DepositInfo(depositParams models.DepositParams) (depositInfo models.DepositInfo, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "deposit"
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), depositParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	err = json.Unmarshal(payload, &depositInfo)
	if err != nil {
		return
	}
	return
}

func (a *AdrestiaRequests) Trade(tradeParams hestia.Trade) (txId string, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "/trade"
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), tradeParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	txId = string(payload)
	txId = strings.ReplaceAll(txId, "\"", "")
	return
}

func (a *AdrestiaRequests) GetTradeStatus(tradeParams hestia.Trade) (tradeInfo hestia.ExchangeOrderInfo, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "/trade/status"
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), tradeParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		log.Println("ERROR::GetTradeStatus::Unmarshall of token string::", string(tokenResponse))
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		log.Println("ERROR::GetTradeStatus:: no header signature")
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	err = json.Unmarshal(payload, &tradeInfo)
	if err != nil {
		log.Println("ERROR::GetTradeStatus::Unmarshall of payload::", payload)
		return
	}
	return
}

func (a *AdrestiaRequests) Withdraw(withdrawParams models.WithdrawParamsV2) (withdrawal models.WithdrawInfo, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "/v2/withdraw"
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), withdrawParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	err = json.Unmarshal(payload, &withdrawal)
	if err != nil {
		return
	}
	return
}

func (a *AdrestiaRequests) GetWithdrawalTxHash(withdrawParams models.WithdrawInfo) (txId string, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "/withdraw/hash"
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), withdrawParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	txId = string(payload)
	txId = strings.ReplaceAll(txId, "\"", "")
	return
}

func (a *AdrestiaRequests) GetPath(fromCoin string, amount float64) (path models.VoucherPathResponse, err error) {
	url := os.Getenv(a.AdrestiaUrl) + "/v2/voucher/path" // this rout internally redirects to path2
	pathParams := models.VoucherPathParamsV2{
		FromCoin: fromCoin,
		AmountEuro: amount,
	}
	req, err := mvt.CreateMVTToken("POST", url, "ladon", os.Getenv("MASTER_PASSWORD"), pathParams, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return
	}
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		err = errors.New("no header signature")
		return
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("ADRESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return
	}
	err = json.Unmarshal(payload, &path)
	if err != nil {
		return
	}
	return
}
