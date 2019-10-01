package services

import (
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

type BitcouService struct {
	BitcouURL    string
	BitcouToken  string
	VouchersList VouchersData
}

type VouchersData struct {
	List        []models.Voucher
	LastUpdated time.Time
}

func (bs *BitcouService) GetVouchersList() ([]models.Voucher, error) {
	if time.Now().Unix() < bs.VouchersList.LastUpdated.Unix()+UpdateVouchersTimeFrame {
		return bs.VouchersList.List, nil
	}
	list, err := bs.getVouchersList()
	if err != nil {
		return nil, err
	}
	bs.VouchersList.List = list
	bs.VouchersList.LastUpdated = time.Now()
	return bs.VouchersList.List, nil
}

func (bs *BitcouService) GetPhoneTopUpList(countryCode string) (interface{}, error) {
	url := os.Getenv("BITCOU_URL") + "voucher/availableVouchersByPhoneNb?phone_number=" + countryCode
	token := "Bearer " + os.Getenv("BITCOU_TOKEN")
	req, err := http.NewRequest("GET", url, nil)
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
	// TODO get response modeled correctly
	var response interface{}
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (bs *BitcouService) getVouchersList() ([]models.Voucher, error) {
	url := os.Getenv("BITCOU_URL") + "voucher/availableVouchers/"
	token := "Bearer " + os.Getenv("BITCOU_TOKEN")
	req, err := http.NewRequest("GET", url, nil)
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
	var response models.BitcouVouchers
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func InitService() *BitcouService {
	service := &BitcouService{
		BitcouURL:   os.Getenv("BITCOU_URL"),
		BitcouToken: os.Getenv("BITCOU_TOKEN"),
		VouchersList: VouchersData{
			List:        []models.Voucher{},
			LastUpdated: time.Time{},
		},
	}
	return service
}
