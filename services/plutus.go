package services

import (
	"encoding/json"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func GetNewPaymentAddress(coin string) (addr string, err error) {
	req, err := mvt.CreateMVTToken("GET", plutus.ProductionURL+"/address/"+coin, "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return addr, err
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return addr, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return addr, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return addr, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return addr, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return addr, err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return addr, err
	}
	return response, nil
}

func SubmitPayment(body plutus.SendAddressBodyReq) (txid string, err error) {
	req, err := mvt.CreateMVTToken("POST", plutus.ProductionURL+"/send/address", "ladon", os.Getenv("MASTER_PASSWORD"), body, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return txid, err
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return txid, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return txid, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return txid, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return txid, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return txid, err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return txid, err
	}
	return response, nil
}

func ValidateRawTx(body plutus.ValidateRawTxReq) (valid bool, err error) {
	req, err := mvt.CreateMVTToken("POST", plutus.ProductionURL+"/validate/tx", "ladon", os.Getenv("MASTER_PASSWORD"), body, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return false, err
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return false, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return false, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return false, err
	}
	var response bool
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return false, err
	}
	return response, nil
}
