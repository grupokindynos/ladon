package processor

import (
	"encoding/json"
	"errors"
	"fmt"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/grupokindynos/common/tokens/mrt"
	"github.com/grupokindynos/common/tokens/mvt"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func Start() {
	fmt.Println("Starting Voucher Processor")
	var wg sync.WaitGroup
	wg.Add(3)
	go handlePendingVouchers(&wg)
	go handleConfirmingVouchers(&wg)
	go handleConfirmedVouchers(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")
}

func handlePendingVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getPendingVouchers()
	if err != nil {
		fmt.Println("Pending vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, v := range vouchers {
		if v.Timestamp+7200 < time.Now().Unix() {
			v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
			_, err = services.UpdateVoucher(v)
			if err != nil {
				fmt.Println("Unable to update voucher confirmations: " + err.Error())
				continue
			}
			continue
		}
		// TODO validate txs

		// Once the tx are fully validated we broadcast the information and mark the voucher as confirming
		paymentCoinConfig, err := coinfactory.GetCoin(v.PaymentData.Coin)
		if err != nil {
			fmt.Println("Unable to get payment coin configuration: " + err.Error())
			continue
		}
		feeCoinConfig, err := coinfactory.GetCoin(v.PaymentData.Coin)
		if err != nil {
			fmt.Println("Unable to get fee coin configuration: " + err.Error())
			continue
		}
		txidFee, err := broadCastTx(feeCoinConfig, v.FeePayment.RawTx)
		if err != nil {
			fmt.Println("Unable to broadcast fee rawTx: " + err.Error())
			continue
		}
		v.FeePayment.Txid = txidFee
		txidPayment, err := broadCastTx(paymentCoinConfig, v.PaymentData.RawTx)
		if err != nil {
			fmt.Println("Unable to broadcast payment rawTx: " + err.Error())
			// TODO if this happens, we need to refund fee payment.
			continue
		}
		v.PaymentData.Txid = txidPayment
		v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusConfirming)
		_, err = services.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher: " + err.Error())
			continue
		}
	}
}

func handleConfirmedVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getConfirmedVouchers()
	if err != nil {
		fmt.Println("Confirmed vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, v := range vouchers {
		txid, err := submitBitcouPayment(v.BitcouPaymentData.Coin, v.BitcouPaymentData.Address, v.BitcouPaymentData.Amount)
		if err != nil {
			// TODO handle refund
			fmt.Println("Unable to submit bitcou payment, should refund the user: " + err.Error())
			continue
		}
		v.BitcouPaymentData.Txid = txid
		_, err = services.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher bitcou payment: " + err.Error())
			continue
		}
	}
}

func handleConfirmingVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getConfirmingVouchers()
	if err != nil {
		fmt.Println("Confirming vouchers processor finished with errors: " + err.Error())
		return
	}
	// Check confirmations and return
	for _, v := range vouchers {
		paymentCoinConfig, err := coinfactory.GetCoin(v.PaymentData.Coin)
		if err != nil {
			fmt.Println("Unable to get payment coin configuration: " + err.Error())
			continue
		}
		feeCoinConfig, err := coinfactory.GetCoin(v.FeePayment.Coin)
		if err != nil {
			fmt.Println("Unable to get fee coin configuration: " + err.Error())
			continue
		}
		// Check if voucher has enough confirmations
		if v.PaymentData.Confirmations >= int32(paymentCoinConfig.BlockchainInfo.MinConfirmations) && v.FeePayment.Confirmations >= int32(feeCoinConfig.BlockchainInfo.MinConfirmations) {
			v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusConfirmed)
			continue
		}
		paymentConfirmations, err := getConfirmations(paymentCoinConfig, v.PaymentData.Txid)
		if err != nil {
			fmt.Println("Unable to get payment coin confirmations: " + err.Error())
			continue
		}
		feeConfirmations, err := getConfirmations(feeCoinConfig, v.FeePayment.Txid)
		if err != nil {
			fmt.Println("Unable to get fee coin confirmations: " + err.Error())
			continue
		}
		v.PaymentData.Confirmations = int32(paymentConfirmations)
		v.FeePayment.Confirmations = int32(feeConfirmations)
		_, err = services.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher confirmations: " + err.Error())
			continue
		}
	}
}

func getPendingVouchers() ([]hestia.Voucher, error) {
	req, err := mvt.CreateMVTToken("GET", hestia.ProductionURL+"/voucher/all?filter="+hestia.GetVoucherStatusString(hestia.VoucherStatusPending), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
	var response []hestia.Voucher
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getConfirmingVouchers() ([]hestia.Voucher, error) {
	req, err := mvt.CreateMVTToken("GET", hestia.ProductionURL+"/voucher/all?filter="+hestia.GetVoucherStatusString(hestia.VoucherStatusConfirming), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
	var response []hestia.Voucher
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getConfirmedVouchers() ([]hestia.Voucher, error) {
	req, err := mvt.CreateMVTToken("GET", hestia.ProductionURL+"/voucher/all?filter="+hestia.GetVoucherStatusString(hestia.VoucherStatusConfirmed), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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
	var response []hestia.Voucher
	err = json.Unmarshal(payload, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getConfirmations(coinConfig *coins.Coin, txid string) (int, error) {
	resp, err := http.Get(coinConfig.BlockExplorer + "/api/v1/tx/" + txid)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var response models.BlockbookTxInfo
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	return response.Confirmations, nil
}

func submitBitcouPayment(coin string, address string, amount int64) (txid string, err error) {
	floatAmount := float64(amount * 1 / 1e8)
	payment := plutus.SendAddressBodyReq{
		Address: address,
		Coin:    coin,
		Amount:  floatAmount,
	}
	return services.SubmitPayment(payment)
}

func broadCastTx(coinConfig *coins.Coin, rawTx string) (txid string, err error) {
	resp, err := http.Get(coinConfig.BlockExplorer + "/api/v2/sendtx/" + rawTx)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var response models.BlockbookBroadcastResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Result, nil
}
