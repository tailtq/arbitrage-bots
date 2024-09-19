package commands

import (
	"arbitrage-bot/helpers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"arbitrage-bot/services/web3"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

type FetchUniswapPoolDataCommand struct {
	web3Service *web3.UniswapWeb3Service
}

// NewFetchUniswapPoolDataCommand ... creates a new FetchUniswapPoolDataCommand
func NewFetchUniswapPoolDataCommand() *FetchUniswapPoolDataCommand {
	return &FetchUniswapPoolDataCommand{
		web3Service: web3.NewUniswapWeb3Service(),
	}
}

// fetchSymbolsFromNetwork ... fetches symbols from the network
func (c *FetchUniswapPoolDataCommand) fetchSymbols(poolData []map[string]string) []*sourceprovider.Symbol {
	var symbols []*sourceprovider.Symbol
	var concurrency = 5
	var channel = make(chan common.Address)
	var wg = sync.WaitGroup{}
	wg.Add(concurrency)

	// Fetch pool data concurrently (5 coroutines)
	for range concurrency {
		go func() {
			defer wg.Done()
			for poolAddress := range channel {
				var symbol = c.web3Service.GetPoolData(poolAddress)
				symbols = append(symbols, &symbol)
			}
		}()
	}
	for _, pair := range poolData {
		var poolAddress = common.HexToAddress(pair["address"])
		channel <- poolAddress
	}
	close(channel)
	wg.Wait()

	return symbols
}

// Fetch ... fetches Uniswap pool data & find triangular pairs
func (c *FetchUniswapPoolDataCommand) Fetch(poolDataTempFilepath string) {
	var poolData []map[string]string
	var err = jsonHelper.ReadJSONFile(poolDataTempFilepath, &poolData)
	helpers.Panic(err)
	// Fetch symbols from the network
	var symbols = c.fetchSymbols(poolData)
	// Find triangular pairs & save to cache
	var sourceProvider = dex.NewUniswapSourceProviderService()
	var triangularPairFinder = arbitrage.TriangularPairFinder{}
	triangularPairs := triangularPairFinder.Handle(symbols)
	err = jsonHelper.WriteJSONFile(sourceProvider.GetArbitragePairCachePath(), triangularPairs)
	helpers.Panic(err)
}
