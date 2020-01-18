package services

import (
	"encoding/json"
	"errors"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type HestiaRequests struct {
	HestiaURL string
}

func (h *HestiaRequests) GetVouchersStatus() (hestia.Config, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/config", "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return hestia.Config{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return hestia.Config{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return hestia.Config{}, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return hestia.Config{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return hestia.Config{}, errors.New("no header signature")
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return hestia.Config{}, err
	}
	var response hestia.Config
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.Config{}, err
	}
	return response, nil
}

func (h *HestiaRequests) GetCoinsConfig() ([]hestia.Coin, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/coins", "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return nil, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return nil, errors.New("no header signature")
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return nil, err
	}
	var response []hestia.Coin
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaRequests) GetVoucherInfo(voucherid string) (hestia.Voucher, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher/single/"+voucherid, "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return hestia.Voucher{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return hestia.Voucher{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return hestia.Voucher{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return hestia.Voucher{}, errors.New("no header signature")
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return hestia.Voucher{}, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return hestia.Voucher{}, err
	}
	var response hestia.Voucher
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.Voucher{}, err
	}
	return response, nil
}

func (h *HestiaRequests) UpdateVoucher(voucherData hestia.Voucher) (string, error) {
	req, err := mvt.CreateMVTToken("POST", os.Getenv(h.HestiaURL)+"/voucher", "ladon", os.Getenv("MASTER_PASSWORD"), voucherData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return "", err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return "", err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return "", err
	}
	var response string
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (h *HestiaRequests) GetVouchersByTimestamp(uid string, timestamp string) (vouchers []hestia.Voucher, err error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher/all_by_timestamp?timestamp="+timestamp+"&userid="+uid, "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return vouchers, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return vouchers, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return vouchers, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return vouchers, err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return vouchers, err
	}
	err = json.Unmarshal(payload, &vouchers)
	if err != nil {
		return vouchers, err
	}
	return vouchers, nil
}
