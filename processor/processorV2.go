package processor

import (
	"fmt"
	amodels "github.com/grupokindynos/adrestia-go/models"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/ladon/models"
	"github.com/grupokindynos/ladon/services"
	"log"
	"strconv"
	"sync"
	"time"
)

type ProcessorV2 struct {
	SkipValidations bool
	Hestia          services.HestiaService
	Plutus          services.PlutusService
	Bitcou          services.BitcouService
	Adrestia        services.AdrestiaService
	HestiaUrl       string
}

func (p *ProcessorV2) Start() {
	var wg sync.WaitGroup
	fmt.Println("Starting ProcessorV2")
	wg.Add(5)
	go p.handlePaymentProcessing(&wg)
	go p.handleRedeemed(&wg)
	go p.handlePerformingTrade(&wg)
	go p.handleNeedsRefund(&wg)
	go p.handleWaitingRefundTxId(&wg)
	fmt.Println("Ending ProcessorV2")
	wg.Wait()
}

func (p *ProcessorV2) handlePaymentProcessing(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers := p.getVoucherByStatus(hestia.VoucherStatusV2PaymentProcessing)
	fmt.Println("Vouchers handlePaymentProcessing ", GetVoucherIds(vouchers))
	for _, voucher := range vouchers {
		coinInfo, err := coinfactory.GetCoin(voucher.UserPayment.Coin)
		if err != nil {
			log.Println("ProcessorV2::handlePaymentProcessing::GetCoin::", voucher.UserPayment.Coin, err.Error())
			continue
		}
		err = checkTxId(&voucher.UserPayment)
		if err != nil {
			log.Println("ProcessorV2::handlePaymentProcessing::checkTxId::", voucher.Id, " ", voucher.UserPayment, " ", err.Error())
			continue
		}
		confirmations, err := getConfirmations(coinInfo, voucher.UserPayment.Txid)
		if err != nil {
			log.Println("handlePaymentProcessing - getConfirmations - ", voucher.Id, " ", err.Error())
			continue
		}
		// log.Println("SIMULATING BITCOU PROCESS FOR ", voucher.Id)
		if confirmations >= coinInfo.BlockchainInfo.MinConfirmations {
			/*res := models.PurchaseInfoResponseV2{
				TxId:       "TEST-TXID",
				AmountEuro: "3",
				RedeemData: "YOU WILL RECEIVE THIS CODE VIA TBD",
			}
			err = nil*/
			res, err := p.Bitcou.GetTransactionInformationV2(models.PurchaseInfo{
				TransactionID: voucher.Id,
				ProductID:     int32(voucher.VoucherId),
				VariantID:     int32(voucher.VariantId),
				PhoneNB:       voucher.PhoneNumber,
				Email:         voucher.Email,
				KYC:           false,
			})
			if err != nil {
				log.Println("handlePaymentProcessing - GetTransactionInformation - " + err.Error())
				voucher.RedeemCode = err.Error()
				voucher.Status = hestia.VoucherStatusV2NeedsRefund
			} else {
				amountEuro, _ := strconv.ParseInt(res.AmountEuro, 10, 64)
				voucher.BitcouTxId = res.TxId
				voucher.RedeemCode = res.RedeemData
				voucher.AmountEuro = amountEuro
				voucher.FulfilledTime = time.Now().Unix()
				voucher.Status = hestia.VoucherStatusV2Redeemed
			}
			_, err = p.Hestia.UpdateVoucherV2(voucher)
			if err != nil {
				log.Println("handlePaymentProcessing - UpdateVoucherV2 - " + err.Error())
			}
		}
	}
}

func (p *ProcessorV2) handleRedeemed(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers := p.getVoucherByStatus(hestia.VoucherStatusV2Redeemed)
	for _, voucher := range vouchers {
		res, err := p.Adrestia.DepositInfo(amodels.DepositParams{
			Asset:   voucher.UserPayment.Coin,
			TxId:    voucher.UserPayment.Txid,
			Address: voucher.UserPayment.Address,
		})
		if err != nil {
			log.Println(voucher.Id)
			log.Println("ProcessorV2::handleRedeemed::DepositInfo::", voucher.UserPayment.Txid, " ", voucher.UserPayment.Coin, " ", voucher.UserPayment.Address, err.Error())
			continue
		}
		if res.DepositInfo.Status == hestia.ExchangeOrderStatusCompleted {
			if voucher.Conversion.Status == hestia.ShiftV2TradeStatusCompleted { // User payed with stable coin.
				voucher.Status = hestia.VoucherStatusV2Complete
			} else {
				voucher.Conversion.Status = hestia.ShiftV2TradeStatusCreated
				voucher.Status = hestia.VoucherStatusV2PerformingTrade
			}

			voucher.Conversion.Conversions[0].Amount = res.DepositInfo.ReceivedAmount
			voucher.ReceivedAmount = res.DepositInfo.ReceivedAmount // Esto se va a sobreescribir si se necesitan trades
			_, err := p.Hestia.UpdateVoucherV2(voucher)
			if err != nil {
				log.Println("ProcessorV2::handleRedeemed::UpdateVoucherV2::" + err.Error())
			}
		}
	}
}

func (p *ProcessorV2) handlePerformingTrade(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers := p.getVoucherByStatus(hestia.VoucherStatusV2PerformingTrade)
	for _, voucher := range vouchers {
		switch voucher.Conversion.Status {
		case hestia.ShiftV2TradeStatusCreated:
			p.handleDirectionalTradeCreated(&voucher.Conversion, voucher.Id)
		case hestia.ShiftV2TradeStatusPerforming:
			p.handleDirectionalTradePerforming(&voucher)
			if voucher.Conversion.Status == hestia.ShiftV2TradeStatusCompleted {
				voucher.Status = hestia.VoucherStatusV2Complete
				voucher.FulfilledTime = time.Now().Unix()
			}
		}
		_, err := p.Hestia.UpdateVoucherV2(voucher)
		if err != nil {
			log.Println("handlePerformingTrade - UpdateVoucherV2 - " + err.Error())
		}
	}
}

func (p *ProcessorV2) handleNeedsRefund(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers := p.getVoucherByStatus(hestia.VoucherStatusV2NeedsRefund)
	for _, voucher := range vouchers {
		res, err := p.Adrestia.DepositInfo(amodels.DepositParams{
			Asset:   voucher.UserPayment.Coin,
			TxId:    voucher.UserPayment.Txid,
			Address: voucher.UserPayment.Address,
		})
		if err != nil {
			log.Println("handleNeedsRefund - DepositInfo - " + err.Error())
			continue
		}
		if res.DepositInfo.Status == hestia.ExchangeOrderStatusCompleted {
			res, err := p.Adrestia.Withdraw(amodels.WithdrawParams{
				Asset:   voucher.UserPayment.Coin,
				Address: voucher.RefundAddress,
				Amount:  res.DepositInfo.ReceivedAmount,
			})
			if err != nil {
				log.Println("handleNeedsRefund - Withdraw - " + err.Error())
				continue
			}
			voucher.RefundTxId = res.TxId
			voucher.Status = hestia.VoucherStatusV2WaitingRefundTxId
			_, err = p.Hestia.UpdateVoucherV2(voucher)
			if err != nil {
				log.Println("handleNeedsRefund - UpdateVoucherV2 - " + err.Error())
			}
		}
	}
}

func (p *ProcessorV2) handleWaitingRefundTxId(wg *sync.WaitGroup) {
	defer wg.Done()
	vouchers := p.getVoucherByStatus(hestia.VoucherStatusV2WaitingRefundTxId)
	for _, voucher := range vouchers {
		txId, err := p.Adrestia.GetWithdrawalTxHash(amodels.WithdrawInfo{
			Exchange: voucher.Conversion.Exchange,
			Asset:    voucher.UserPayment.Coin,
			TxId:     voucher.RefundTxId,
		})
		if err != nil {
			log.Println("handleWaitingRefundTxId - GetWithdrawalTxHash - " + err.Error())
			continue
		}
		if txId != "" {
			voucher.RefundTxId = txId
			voucher.FulfilledTime = time.Now().Unix()
			voucher.Status = hestia.VoucherStatusV2Refunded
			_, err := p.Hestia.UpdateVoucherV2(voucher)
			if err != nil {
				log.Println("handleWaitingRefundTxId - UpdateVoucherV2 - " + err.Error())
			}
		}
	}
}

func (p *ProcessorV2) getVoucherByStatus(status hestia.VoucherStatusV2) (vouchers []hestia.VoucherV2) {
	var err error
	vouchers, err = p.Hestia.GetVouchersByStatusV2(status)
	if err != nil {
		log.Println("Unable to get vouchers with status " + hestia.GetVoucherStatusV2String(status))
	}
	log.Println(hestia.GetVoucherStatusV2String(status), vouchers)
	return
}

func (p *ProcessorV2) handleDirectionalTradePerforming(voucher *hestia.VoucherV2) {
	dt := &voucher.Conversion
	if dt.Conversions[0].Status == hestia.ExchangeOrderStatusCompleted {
		res, err := p.Adrestia.GetTradeStatus(dt.Conversions[1])
		if err != nil {
			log.Println("handleDirectionalTradePerforming - GetTradeStatus1 - " + err.Error())
			return
		}
		if res.Status == hestia.ExchangeOrderStatusCompleted {
			dt.Conversions[1].ReceivedAmount = res.ReceivedAmount
			dt.Conversions[1].FulfilledTime = time.Now().Unix()
			dt.Conversions[1].Status = hestia.ExchangeOrderStatusCompleted
			dt.Status = hestia.ShiftV2TradeStatusCompleted
			voucher.ReceivedAmount = res.ReceivedAmount
		}
	} else {
		res, err := p.Adrestia.GetTradeStatus(dt.Conversions[0])
		if err != nil {
			log.Println("handleDirectionalTradePerforming - GetTradeStatus2 - " + err.Error())
			return
		}
		if res.Status == hestia.ExchangeOrderStatusCompleted {
			if len(dt.Conversions) > 1 {
				dt.Conversions[1].Amount = res.ReceivedAmount
				res, err := p.Adrestia.Trade(dt.Conversions[1])
				if err != nil {
					log.Println("handleDirectionalTradePerforming::Trade::", voucher.Id + " " + err.Error())
					return
				}
				dt.Conversions[1].OrderId = res
				dt.Conversions[1].CreatedTime = time.Now().Unix()
				dt.Conversions[1].Status = hestia.ExchangeOrderStatusOpen
			} else {
				dt.Status = hestia.ShiftV2TradeStatusCompleted
				voucher.ReceivedAmount = res.ReceivedAmount
			}
			dt.Conversions[0].Status = hestia.ExchangeOrderStatusCompleted
			dt.Conversions[0].ReceivedAmount = res.ReceivedAmount
			dt.Conversions[0].FulfilledTime = time.Now().Unix()
		}
	}
}

func (p *ProcessorV2) handleDirectionalTradeCreated(dt *hestia.DirectionalTrade, id string) {
	res, err := p.Adrestia.Trade(dt.Conversions[0])
	if err != nil {
		log.Println("handleDirectionalTradeCreated - Trade - " + id, " ", err.Error())
		return
	}

	dt.Conversions[0].Status = hestia.ExchangeOrderStatusOpen
	dt.Conversions[0].CreatedTime = time.Now().Unix()
	dt.Conversions[0].OrderId = res
	dt.Status = hestia.ShiftV2TradeStatusPerforming
}

func GetVoucherIds(hestiaVouchers []hestia.VoucherV2) (vouchers []string){
	for _, voucher := range hestiaVouchers {
		vouchers = append(vouchers, voucher.Id)
	}
	return
}
