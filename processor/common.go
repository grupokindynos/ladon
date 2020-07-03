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
	explorerWrapper, _ := explorer.NewExplorerFactory().GetExplorerByCoin(*coinConfig)
	return explorerWrapper.FindDepositTxId(address, amount)
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
	if coinConfig.Info.Token {
		coinConfig, _ = coinfactory.GetCoin("ETH")
	}
	explorerWrapper, _ := explorer.NewExplorerFactory().GetExplorerByCoin(*coinConfig)
	txData, err := explorerWrapper.GetTx(txid)
	if err != nil {
		return 0, err
	}
	return txData.Confirmations, nil
}
