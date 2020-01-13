package services

import (
	"bytes"
	"encoding/json"
	"github.com/grupokindynos/ladon/models"
	"io/ioutil"
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
}

type VouchersData struct {
	List        map[string][]models.Voucher
	LastUpdated time.Time
}

func (bs *BitcouRequests) GetPhoneTopUpList(phoneNb string) ([]int, error) {
	url := os.Getenv("BITCOU_URL") + "voucher/availableVouchersByPhoneNb"
	token := "Bearer " + os.Getenv("BITCOU_TOKEN")
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

func (bs *BitcouRequests) GetTransactionInformation(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponse, error) {
	url := os.Getenv("BITCOU_URL") + "voucher/transaction"
	token := "Bearer " + os.Getenv("BITCOU_TOKEN")
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

func NewBitcouService() *BitcouRequests {
	service := &BitcouRequests{
		BitcouURL:   os.Getenv("BITCOU_URL"),
		BitcouToken: os.Getenv("BITCOU_TOKEN"),
		VouchersList: VouchersData{
			List:        make(map[string][]models.Voucher),
			LastUpdated: time.Time{},
		},
	}
	return service
}
