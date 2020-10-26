package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/ladon/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	UpdateVouchersTimeFrame = 60 * 60 * 24 // 1 day
)

type BitcouRequests struct {
	BitcouURL    string
	BitcouToken  string
	VouchersList VouchersData
	DevMode      bool
}

type VouchersData struct {
	List        map[string][]models.Voucher
	LastUpdated time.Time
}

func (bs *BitcouRequests) GetPhoneTopUpList(phoneNb string) ([]int, error) {
	url := bs.BitcouURL + "voucher/availableVouchersByPhoneNb"
	token := "Bearer " + bs.BitcouToken

	body := models.BitcouPhoneBodyReq{PhoneNumber: phoneNb}
	byteBody, err := json.Marshal(body)
	postBody := bytes.NewBuffer(byteBody)
	req, err := http.NewRequest("POST", url, postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response models.BitcouPhoneResponseList
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return nil, err
	}
	var productIDs []int
	for _, product := range response.Data {
		productIDs = append(productIDs, product.ProductID)
	}
	return productIDs, nil
}

func (bs *BitcouRequests) GetPhoneTopUpListV2(phoneNb string) ([]int, error) {
	url := bs.BitcouURL + "voucher/availableVouchersByPhoneNb"
	token := "Bearer " + bs.BitcouToken
	body := models.BitcouPhoneBodyReq{PhoneNumber: phoneNb}
	byteBody, err := json.Marshal(body)
	postBody := bytes.NewBuffer(byteBody)
	req, err := http.NewRequest("POST", url, postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response models.BitcouPhoneResponseList
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return nil, err
	}
	var productIDs []int
	for _, product := range response.Data {
		productIDs = append(productIDs, product.ProductID)
	}
	return productIDs, nil
}


func (bs *BitcouRequests) GetTransactionInformation(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponse, error) {
	url := bs.BitcouURL + "voucher/transaction"
	token := "Bearer " + bs.BitcouToken

	byteBody, err := json.Marshal(purchaseInfo)
	postBody := bytes.NewBuffer(byteBody)
	req, err := http.NewRequest("POST", url, postBody)
	if err != nil {
		return models.PurchaseInfoResponse{}, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return models.PurchaseInfoResponse{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response models.BitcouBaseResponse
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return models.PurchaseInfoResponse{}, err
	}
	var purchaseData models.PurchaseInfoResponse
	dataBytes, err := json.Marshal(response.Data[0])
	if err != nil {
		return models.PurchaseInfoResponse{}, err
	}
	err = json.Unmarshal(dataBytes, &purchaseData)
	if err != nil {
		return models.PurchaseInfoResponse{}, err
	}
	return purchaseData, nil
}

func (bs *BitcouRequests) GetTransactionInformationV2(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponseV2, error) {
	url := bs.BitcouURL + "voucher/transaction"
	token := "Bearer " + bs.BitcouToken

	byteBody, err := json.Marshal(purchaseInfo)
	postBody := bytes.NewBuffer(byteBody)
	req, err := http.NewRequest("POST", url, postBody)
	if err != nil {
		return models.PurchaseInfoResponseV2{}, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return models.PurchaseInfoResponseV2{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response models.BitcouBaseResponse
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return models.PurchaseInfoResponseV2{}, err
	}
	var purchaseData models.PurchaseInfoResponseV2
	if len(response.Data) >= 1 {
		dataBytes, err := json.Marshal(response.Data[0])
		if err != nil {
			return models.PurchaseInfoResponseV2{}, err
		}
		err = json.Unmarshal(dataBytes, &purchaseData)
		if err != nil {
			return models.PurchaseInfoResponseV2{}, err
		}
		return purchaseData, nil
	} else {
		log.Println("GetTransactionInformationV2:: bad response", string(contents))
		return purchaseData, errors.New("bad response from Bitcou::" + string(contents))
	}
}

func (bs *BitcouRequests) GetAccountBalanceV2() (models.AccountInfo, error) {
	url := bs.BitcouURL + "account/balance"
	token := "Bearer " + bs.BitcouToken
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.AccountInfo{}, err
	}
	req.Header.Add("Authorization", token)
	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return models.AccountInfo{}, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	contents, _ := ioutil.ReadAll(res.Body)
	var response models.BitcouBaseResponse
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return models.AccountInfo{}, err
	}
	var purchaseData models.AccountInfo
	if len(response.Data) >= 1 {
		dataBytes, err := json.Marshal(response.Data[0])
		if err != nil {
			return models.AccountInfo{}, err
		}
		err = json.Unmarshal(dataBytes, &purchaseData)
		if err != nil {
			return models.AccountInfo{}, err
		}
		return purchaseData, nil
	} else {
		log.Println("GetTransactionInformationV2:: bad response", string(contents))
		return purchaseData, errors.New("bad response from Bitcou")
	}
}

func NewBitcouService(devMode bool, version int) *BitcouRequests {
	var url string
	if devMode {
		if version == 1 {
			url = os.Getenv("BITCOU_DEV_URL_V1")
		} else {
			url = os.Getenv("BITCOU_DEV_URL_V2")
		}
	} else {
		if version == 1 {
			url = os.Getenv("BITCOU_URL_V1")
		} else {
			url = os.Getenv("BITCOU_URL_V2")
		}
	}

	var token string
	if version == 1 {
		token = os.Getenv("BITCOU_TOKEN_V1")
	} else {
		token = os.Getenv("BITCOU_TOKEN_V2")
	}

	service := &BitcouRequests{
		BitcouURL:   url,
		BitcouToken: token,
		VouchersList: VouchersData{
			List:        make(map[string][]models.Voucher),
			LastUpdated: time.Time{},
		},
		DevMode: devMode,
	}

	fmt.Print("Using Bitcou URL: ", service.BitcouURL, service.BitcouToken)
	return service
}