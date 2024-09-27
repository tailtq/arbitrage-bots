package web3

import (
	sp "arbitrage-bot/services/sourceprovider"
	"sync"
)

type DEXWeb3Service interface {
	GetPrice(symbol sp.Symbol, amountIn float64, tradeDirection string, verbose bool) float64
	GetPriceMultiplePaths(tradePaths []sp.TradePath, amountIn float64, verbose bool) float64
	AggregatePrices(symbols []*sp.Symbol, verbose bool) *sync.Map
}
