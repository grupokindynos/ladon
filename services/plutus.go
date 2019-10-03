package services

import (
	"encoding/json"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// TODO use common module
var PlutusURL = "https://plutus.polispay.com"

func GetNewPaymentAddress(coin string) (addr string, err error) {
	req, err := mvt.CreateMVTToken("GET", "http://localhost:8082/address/"+coin, "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
