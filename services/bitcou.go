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

type BitcouService struct {
	BitcouURL    string
	BitcouToken  string
	VouchersList VouchersData
}

type VouchersData struct {
	List        map[string][]models.Voucher
	LastUpdated time.Time
}

func (bs *BitcouService) GetVouchersList() (map[string][]models.Voucher, error) {
	if time.Now().Unix() < bs.VouchersList.LastUpdated.Unix()+UpdateVouchersTimeFrame {
		return bs.VouchersList.List, nil
	}
	list, err := bs.getVouchersList()
	if err != nil {
		return nil, err
	}
	VouchersList := make(map[string][]models.Voucher)
	for _, v := range list {
		if v.Countries.Austria {
			VouchersList["austria"] = append(VouchersList["austria"], v)
		}
		if v.Countries.Belgium {
			VouchersList["belgium"] = append(VouchersList["belgium"], v)
		}
		if v.Countries.Bulgaria {
			VouchersList["bulgaria"] = append(VouchersList["bulgaria"], v)
		}
		if v.Countries.Canada {
			VouchersList["canada"] = append(VouchersList["canada"], v)
		}
		if v.Countries.Croatia {
			VouchersList["croatia"] = append(VouchersList["croatia"], v)
		}
		if v.Countries.Cyprus {
			VouchersList["cyprus"] = append(VouchersList["cyprus"], v)
		}
		if v.Countries.Czechia {
			VouchersList["czechia"] = append(VouchersList["czechia"], v)
		}
		if v.Countries.Denmark {
			VouchersList["denmark"] = append(VouchersList["denmark"], v)
		}
		if v.Countries.Estonia {
			VouchersList["estonia"] = append(VouchersList["estonia"], v)
		}
		if v.Countries.Finland {
			VouchersList["finland"] = append(VouchersList["finland"], v)
		}
		if v.Countries.France {
			VouchersList["france"] = append(VouchersList["france"], v)
		}
		if v.Countries.Germany {
			VouchersList["germany"] = append(VouchersList["germany"], v)
		}
		if v.Countries.GreatBritain {
			VouchersList["great_britain"] = append(VouchersList["great_britain"], v)
		}
		if v.Countries.Greece {
			VouchersList["greece"] = append(VouchersList["greece"], v)
		}
		if v.Countries.Hongkong {
			VouchersList["hongkond"] = append(VouchersList["hongkond"], v)
		}
		if v.Countries.Hungary {
			VouchersList["hungary"] = append(VouchersList["hungary"], v)
		}
		if v.Countries.Indonesia {
			VouchersList["indonesia"] = append(VouchersList["indonesia"], v)
		}
		if v.Countries.Ireland {
			VouchersList["ireland"] = append(VouchersList["ireland"], v)
		}
		if v.Countries.Italy {
			VouchersList["italy"] = append(VouchersList["italy"], v)
		}
		if v.Countries.Lichtenstein {
			VouchersList["lichtenstein"] = append(VouchersList["lichtenstein"], v)
		}
		if v.Countries.Luxembourg {
			VouchersList["luxembourg"] = append(VouchersList["luxembourg"], v)
		}
		if v.Countries.Malaysia {
			VouchersList["malaysia"] = append(VouchersList["malaysia"], v)
		}
		if v.Countries.Malta {
			VouchersList["malta"] = append(VouchersList["malta"], v)
		}
		if v.Countries.Mexico {
			VouchersList["mexico"] = append(VouchersList["mexico"], v)
		}
		if v.Countries.Netherland {
			VouchersList["netherland"] = append(VouchersList["netherland"], v)
		}
		if v.Countries.Norway {
			VouchersList["norway"] = append(VouchersList["norway"], v)
		}
		if v.Countries.Poland {
			VouchersList["poland"] = append(VouchersList["poland"], v)
		}
		if v.Countries.Portugal {
			VouchersList["portugal"] = append(VouchersList["portugal"], v)
		}
		if v.Countries.Russia {
			VouchersList["russia"] = append(VouchersList["russia"], v)
		}
		if v.Countries.Singapore {
			VouchersList["singapore"] = append(VouchersList["singapore"], v)
		}
		if v.Countries.Slovakia {
			VouchersList["slovakia"] = append(VouchersList["slovakia"], v)
		}
		if v.Countries.Slovenia {
			VouchersList["slovenia"] = append(VouchersList["slovenia"], v)
		}
		if v.Countries.Spain {
			VouchersList["spain"] = append(VouchersList["spain"], v)
		}
		if v.Countries.Sweden {
			VouchersList["sweden"] = append(VouchersList["sweden"], v)
		}
		if v.Countries.Switzerland {
			VouchersList["switzerland"] = append(VouchersList["switzerland"], v)
		}
		if v.Countries.Turkey {
			VouchersList["turkey"] = append(VouchersList["turkey"], v)
		}
		if v.Countries.Usa {
			VouchersList["usa"] = append(VouchersList["usa"], v)
		}
	}
	bs.VouchersList.List = VouchersList
	bs.VouchersList.LastUpdated = time.Now()
	return bs.VouchersList.List, nil
}

func (bs *BitcouService) GetPhoneTopUpList(phoneNb string) ([]int, error) {
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

func (bs *BitcouService) GetTransactionInformation(purchaseInfo models.PurchaseInfo) (models.PurchaseInfoResponse, error) {
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
	var response models.BitcouBaseResponse
	err = json.Unmarshal(contents, &response)
	if err != nil {
		return nil, err
	}
	var vouchersList []models.Voucher
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataBytes, &vouchersList)
	if err != nil {
		return nil, err
	}
	return vouchersList, nil
}

func InitService() *BitcouService {
	service := &BitcouService{
		BitcouURL:   os.Getenv("BITCOU_URL"),
		BitcouToken: os.Getenv("BITCOU_TOKEN"),
		VouchersList: VouchersData{
			List:        make(map[string][]models.Voucher),
			LastUpdated: time.Time{},
		},
	}
	return service
}
