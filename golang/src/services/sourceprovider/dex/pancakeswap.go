package dex

import (
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/web3"
	"os"
	"sync"
	"time"
)

type PancakeswapSourceProvider struct {
	web3Service     *web3.PancakeswapWeb3Service
	symbolPriceData sync.Map
	symbols         map[string]*sourceprovider.Symbol
}

func NewPancakeswapSourceProvider() *PancakeswapSourceProvider {
	return &PancakeswapSourceProvider{
		web3Service: web3.NewPancakeswapWeb3Service(),
	}
}

func (p *PancakeswapSourceProvider) Web3Service() *web3.PancakeswapWeb3Service {
	return p.web3Service
}

// GetArbitragePairCachePath ... returns the path to the token list cache
func (p *PancakeswapSourceProvider) GetArbitragePairCachePath() string {
	var network = os.Getenv("NETWORK_NAME")
	return "data/" + network + "/pancakeswapArbitragePairs.json"
}

// GetSymbolPrice ... returns the aggregated price for a given symbol
func (p *PancakeswapSourceProvider) GetSymbolPrice(symbol string) *SymbolPrice {
	if price, ok := p.symbolPriceData.Load(symbol); ok {
		return price.(*SymbolPrice)
	}
	return nil
}

// GetSymbol ... returns the symbol
func (p *PancakeswapSourceProvider) GetSymbol(symbol string) sourceprovider.Symbol {
	return *p.symbols[symbol]
}

// SubscribeSymbols ... subscribes to the symbols
func (p *PancakeswapSourceProvider) SubscribeSymbols(symbols []*sourceprovider.Symbol, pingChannel chan bool) {
	var tokenPairs []string

	for _, symbol := range symbols {
		p.symbols[symbol.Symbol] = symbol
		tokenPairs = append(tokenPairs, symbol.Symbol)
	}

	for {
		aggregatedPrices := p.web3Service.AggregatePrices(symbols, true)
		aggregatedPrices.Range(func(key any, value any) bool {
			p.symbolPriceData.Store(key, &SymbolPrice{
				Symbol:      p.symbols[key.(string)],
				Token0Price: 1.0 / value.(float64),
				Token1Price: value.(float64),
				EventTime:   time.Now(),
			})
			return true
		})
		pingChannel <- true
		// Fetch the data every 60 seconds
		time.Sleep(10 * time.Second)
	}
}
