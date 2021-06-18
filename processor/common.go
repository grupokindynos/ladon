package processor

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/explorer"
	"github.com/grupokindynos/common/hestia"
)

func getMissingTxId(coin string, address string, amount int64) (string, error) {
	coinConfig, _ := coinfactory.GetCoin(coin)
	if coinConfig.Info.Token {
		coinConfig, _ = coinfactory.GetCoin("ETH")
	}
	blockBook := explorer.NewBlockBookWrapper(coinConfig.Info.Blockbook)
	return blockBook.FindDepositTxId(address, amount)
}


func checkTxId(payment *hestia.Payment) error {
	if payment.Txid == "" {
		txId, err := getMissingTxId(payment.Coin, payment.Address, payment.Amount)
		if err != nil {
			return err
		}
		payment.Txid = txId
	}
	return nil
}

func getConfirmations(coinConfig *coins.Coin, txid string) (int, error) {
	explorerWrapper, err := getBlockExplorerWrapper(coinConfig)
	if err != nil {
		return 0, err
	}
	txData, err := explorerWrapper.GetTx(txid)
	if err != nil {
		return 0, err
	}
	return txData.Confirmations, nil
}

func getBlockExplorerWrapper(coinConfig * coins.Coin) (explorer.Explorer, error) {
	var err error
	if coinConfig.Info.Token {
		if coinConfig.Info.TokenNetwork == "ethereum" {
			coinConfig, err = coinfactory.GetCoin("ETH")
			if err != nil {
				return nil, err
			}
		} else if coinConfig.Info.TokenNetwork == "binance" {
			coinConfig, err = coinfactory.GetCoin("BNB")
			if err != nil {
				return nil, err
			}
			return explorer.NewBscScanWrapper("https://api.bscscan.com"), nil
		}
	}
	return explorer.NewBlockBookWrapper(coinConfig.Info.Blockbook), nil
}