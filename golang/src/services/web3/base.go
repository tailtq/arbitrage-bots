package web3

import (
	"arbitrage-bot/services/sourceprovider"
	"sync"
)

type DEXWeb3Service interface {
	GetPrice(symbol sourceprovider.Symbol, amountIn float64, tradeDirection string, verbose bool) float64
	AggregatePrices(symbols []*sourceprovider.Symbol, verbose bool) *sync.Map
}
