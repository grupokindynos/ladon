package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/ladon/models"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type HestiaRequests struct {
	HestiaURL string
	TestingDb bool
}

func (h *HestiaRequests) GetVouchersStatus() (hestia.Config, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/config?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))

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

func (h *HestiaRequests) GetVouchersByStatusV2(status hestia.VoucherStatusV2) ([]hestia.VoucherV2, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher2/all?filter="+fmt.Sprintf("%d", status)+"&test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: 40 * time.Second,
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
	var response []hestia.VoucherV2
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *HestiaRequests) GetCoinsConfig() ([]hestia.Coin, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/coins?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher/single/"+voucherid+"?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func (h *HestiaRequests) GetVoucherV2(voucherid string) (hestia.VoucherV2, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher2/single/"+voucherid+"?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return hestia.VoucherV2{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return hestia.VoucherV2{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return hestia.VoucherV2{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return hestia.VoucherV2{}, errors.New("no header signature")
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return hestia.VoucherV2{}, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return hestia.VoucherV2{}, err
	}
	var response hestia.VoucherV2
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return hestia.VoucherV2{}, err
	}
	return response, nil
}

func (h *HestiaRequests) GetVoucherInfoV2(country string, productId string) (models.LightVoucherV2, error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher2/getVoucherInfo/"+country+"/"+productId+"?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return models.LightVoucherV2{}, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return models.LightVoucherV2{}, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return models.LightVoucherV2{}, err
	}
	headerSignature := res.Header.Get("service")
	if headerSignature == "" {
		return models.LightVoucherV2{}, errors.New("no header signature")
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return models.LightVoucherV2{}, err
	}
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return models.LightVoucherV2{}, err
	}
	var response models.LightVoucherV2
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return models.LightVoucherV2{}, err
	}
	return response, nil
}

func (h *HestiaRequests) UpdateVoucher(voucherData hestia.Voucher) (string, error) {
	req, err := mvt.CreateMVTToken("POST", os.Getenv(h.HestiaURL)+"/voucher?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), voucherData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func (h *HestiaRequests) UpdateVoucherV2(voucherData hestia.VoucherV2) (string, error) {
	req, err := mvt.CreateMVTToken("POST", os.Getenv(h.HestiaURL)+"/voucher2?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), voucherData, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher/all_by_timestamp?timestamp="+timestamp+"&userid="+uid+"test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func (h *HestiaRequests) GetVouchersByTimestampV2(uid string, timestamp string) (vouchers []hestia.VoucherV2, err error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher2/all_by_timestamp?timestamp="+timestamp+"&userid="+uid+"&test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func (h *HestiaRequests) GetUserInfo(uid string) (info string, err error) {
	req, err := mvt.CreateMVTToken("GET", os.Getenv(h.HestiaURL)+"/voucher2/user/info?userid="+uid+"&test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
	if err != nil {
		return info, err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()
	tokenResponse, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return info, err
	}
	var tokenString string
	err = json.Unmarshal(tokenResponse, &tokenString)
	if err != nil {
		return info, err
	}
	headerSignature := res.Header.Get("service")
	valid, payload := mrt.VerifyMRTToken(headerSignature, tokenString, os.Getenv("HESTIA_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
	if !valid {
		return info, err
	}
	err = json.Unmarshal(payload, &info)
	if err != nil {
		return info, err
	}
	return info, nil
}

func (h *HestiaRequests) GetVoucherStatus() (hestia.Config, error) {
	req, err := mvt.CreateMVTToken("GET", h.HestiaURL+"/config?test="+b2s(h.TestingDb), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func b2s(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
