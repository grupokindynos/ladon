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
	"github.com/grupokindynos/olympus-utils/amount"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const timeoutAwaiting = 60 * 60 * 2 // 2 hours.

type Processor struct {
	SkipValidations bool
	Hestia          services.HestiaService
	Plutus          services.PlutusService
}

func (p *Processor) Start() {
	fmt.Println("Starting Voucher Processor")
	voucherStatus, err := p.Hestia.GetVouchersStatus()
	if err != nil {
		panic(err)
	}
	if !voucherStatus.Vouchers.Processor {
		fmt.Println("Voucher processor disabled")
		return
	}
	var wg sync.WaitGroup
	wg.Add(6)
	go p.handlePendingVouchers(&wg)
	go p.handleConfirmingVouchers(&wg)
	go p.handleConfirmedVouchers(&wg)
	go p.handleRefundFeeVouchers(&wg)
	go p.handleRefundTotalVouchers(&wg)
	go p.handleTimeoutAwaitingVouchers(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")
}

func (p *Processor) handlePendingVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getPendingVouchers()
	if err != nil {
		fmt.Println("Pending vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, v := range vouchers {
		if v.Timestamp+timeoutAwaiting < time.Now().Unix() {
			v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusError)
			_, err = p.Hestia.UpdateVoucher(v)
			if err != nil {
				fmt.Println("Unable to update voucher confirmations: " + err.Error())
				continue
			}
			continue
		}
		v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusConfirming)
		_, err = p.Hestia.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher: " + err.Error())
			continue
		}
	}
}

func (p *Processor) handleConfirmedVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getConfirmedVouchers()
	if err != nil {
		fmt.Println("Confirmed vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, v := range vouchers {
		txid, err := p.submitBitcouPayment(v.BitcouPaymentData.Coin, v.BitcouPaymentData.Address, v.BitcouPaymentData.Amount)
		if err != nil {
			v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundTotal)
			_, err = p.Hestia.UpdateVoucher(v)
			if err != nil {
				fmt.Println("Unable to update voucher bitcou payment: " + err.Error())
				continue
			}
			fmt.Println("Unable to submit bitcou payment, should refund the user: " + err.Error())
			continue
		}
		v.BitcouPaymentData.Txid = txid
		v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusAwaitingProvider)
		_, err = p.Hestia.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher bitcou payment: " + err.Error())
			continue
		}
	}
}

func (p *Processor) handleConfirmingVouchers(wg *sync.WaitGroup) {
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
		if v.PaymentData.Coin != "POLIS" {
			feeCoinConfig, err := coinfactory.GetCoin(v.FeePayment.Coin)
			if err != nil {
				fmt.Println("Unable to get fee coin configuration: " + err.Error())
				continue
			}
			// Check if voucher has enough confirmations
			if p.SkipValidations || (v.PaymentData.Confirmations >= int32(paymentCoinConfig.BlockchainInfo.MinConfirmations) && v.FeePayment.Confirmations >= int32(feeCoinConfig.BlockchainInfo.MinConfirmations)) {
				v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusConfirmed)
				_, err = p.Hestia.UpdateVoucher(v)
				if err != nil {
					fmt.Println("Unable to update voucher confirmations: " + err.Error())
					continue
				}
				continue
			}
			feeConfirmations, err := getConfirmations(feeCoinConfig, v.FeePayment.Txid)
			if err != nil {
				fmt.Println("Unable to get fee coin confirmations: " + err.Error())
				continue
			}
			v.FeePayment.Confirmations = int32(feeConfirmations)
		} else {
			if p.SkipValidations || v.PaymentData.Confirmations >= int32(paymentCoinConfig.BlockchainInfo.MinConfirmations) {
				v.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusConfirmed)
				_, err = p.Hestia.UpdateVoucher(v)
				if err != nil {
					fmt.Println("Unable to update voucher confirmations: " + err.Error())
					continue
				}
				continue
			}
		}
		paymentConfirmations, err := getConfirmations(paymentCoinConfig, v.PaymentData.Txid)
		if err != nil {
			fmt.Println("Unable to get payment coin confirmations: " + err.Error())
			continue
		}
		v.PaymentData.Confirmations = int32(paymentConfirmations)
		_, err = p.Hestia.UpdateVoucher(v)
		if err != nil {
			fmt.Println("Unable to update voucher confirmations: " + err.Error())
			continue
		}
	}
}

func (p *Processor) handleRefundFeeVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getRefundFeeVouchers()
	if err != nil {
		fmt.Println("Refund Fee vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, voucher := range vouchers {
		paymentBody := plutus.SendAddressBodyReq{
			Address: voucher.RefundFeeAddr,
			Coin:    "POLIS",
			Amount:  amount.AmountType(voucher.FeePayment.Amount).ToNormalUnit(),
		}
		_, err := p.Plutus.SubmitPayment(paymentBody)
		if err != nil {
			fmt.Println("unable to submit refund payment")
			continue
		}
		voucher.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefunded)
		_, err = p.Hestia.UpdateVoucher(voucher)
		if err != nil {
			fmt.Println("unable to update voucher")
			continue
		}
	}
}

func (p *Processor) handleRefundTotalVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getRefundTotalVouchers()
	if err != nil {
		fmt.Println("Refund Total vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, voucher := range vouchers {
		if voucher.PaymentData.Coin == "POLIS" {
			paymentBody := plutus.SendAddressBodyReq{
				Address: voucher.RefundAddr,
				Coin:    voucher.PaymentData.Coin,
				Amount:  amount.AmountType(voucher.PaymentData.Amount).ToNormalUnit(),
			}
			_, err = p.Plutus.SubmitPayment(paymentBody)
			if err != nil {
				fmt.Println("unable to submit refund payment")
				continue
			}
			voucher.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefunded)
			_, err = p.Hestia.UpdateVoucher(voucher)
			if err != nil {
				fmt.Println("unable to update voucher")
				continue
			}
		} else {
			feePaymentBody := plutus.SendAddressBodyReq{
				Address: voucher.RefundFeeAddr,
				Coin:    "POLIS",
				Amount:  amount.AmountType(voucher.FeePayment.Amount).ToNormalUnit(),
			}
			_, err := p.Plutus.SubmitPayment(feePaymentBody)
			if err != nil {
				fmt.Println("unable to submit refund payment")
				continue
			}
			voucher.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundedPartially)
			_, err = p.Hestia.UpdateVoucher(voucher)
			if err != nil {
				fmt.Println("unable to update voucher")
				continue
			}
			paymentBody := plutus.SendAddressBodyReq{
				Address: voucher.RefundAddr,
				Coin:    voucher.PaymentData.Coin,
				Amount:  amount.AmountType(voucher.PaymentData.Amount).ToNormalUnit(),
			}
			_, err = p.Plutus.SubmitPayment(paymentBody)
			if err != nil {
				fmt.Println("unable to submit refund payment")
				continue
			}
			voucher.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefunded)
			_, err = p.Hestia.UpdateVoucher(voucher)
			if err != nil {
				fmt.Println("unable to update voucher")
				continue
			}
		}
	}
}

func (p *Processor) handleTimeoutAwaitingVouchers(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers, err := getAwaitingProviderVouchers()
	if err != nil {
		fmt.Println("Await provider vouchers processor finished with errors: " + err.Error())
		return
	}
	for _, voucher := range vouchers {
		if voucher.Timestamp+timeoutAwaiting < time.Now().Unix() {
			voucher.Status = hestia.GetVoucherStatusString(hestia.VoucherStatusRefundTotal)
			_, err = p.Hestia.UpdateVoucher(voucher)
			if err != nil {
				fmt.Println("unable to update voucher")
				continue
			}
		}
	}
}

func getPendingVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusPending)
}

func getConfirmingVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusConfirming)
}

func getConfirmedVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusConfirmed)
}

func getRefundTotalVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusRefundTotal)
}

func getRefundFeeVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusRefundFee)
}

func getAwaitingProviderVouchers() ([]hestia.Voucher, error) {
	return getVouchers(hestia.VoucherStatusAwaitingProvider)
}

func getVouchers(status hestia.VoucherStatus) ([]hestia.Voucher, error) {
	req, err := mvt.CreateMVTToken("GET", hestia.ProductionURL+"/voucher/all?filter="+hestia.GetVoucherStatusString(status), "ladon", os.Getenv("MASTER_PASSWORD"), nil, os.Getenv("HESTIA_AUTH_USERNAME"), os.Getenv("HESTIA_AUTH_PASSWORD"), os.Getenv("LADON_PRIVATE_KEY"))
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

func (p *Processor) submitBitcouPayment(coin string, address string, amount int64) (txid string, err error) {
	floatAmount := float64(amount)
	payment := plutus.SendAddressBodyReq{
		Address: address,
		Coin:    coin,
		Amount:  floatAmount / 1e8,
	}
	return p.Plutus.SubmitPayment(payment)
}
