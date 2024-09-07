package commands

import (
	"arbitrage-bot/helpers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"arbitrage-bot/services/web3"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

// fetchSymbolsFromNetwork ... fetches symbols from the network
func fetchSymbolsFromNetwork(poolData []map[string]string) []*sourceprovider.Symbol {
	var symbols []*sourceprovider.Symbol
	var uniswapWeb3Service = web3.NewUniswapWeb3Service()
	var concurrency = 5
	var channel = make(chan common.Address)
	var wg = sync.WaitGroup{}
	wg.Add(concurrency)

	// Fetch pool data concurrently (5 coroutines)
	for range concurrency {
		go func() {
			defer wg.Done()
			for poolAddress := range channel {
				var symbol = uniswapWeb3Service.GetPoolData(poolAddress)
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

// FetchUniswapPoolData ... fetches Uniswap pool data & find triangular pairs
func FetchUniswapPoolData(poolDataTemp string) {
	var poolData []map[string]string
	var err = jsonHelper.ReadJSONFile(poolDataTemp, &poolData)
	helpers.Panic(err)

	// Fetch symbols from the network
	var symbols = fetchSymbolsFromNetwork(poolData)
	fmt.Println("Fetched", len(symbols), "symbols")

	// Find triangular pairs & save to cache
	var sourceProvider = dex.GetSourceProvider(sourceprovider.SourceProviderName["Uniswap"])
	var triangularPairFinder = arbitrage.TriangularPairFinder{}
	triangularPairs := triangularPairFinder.Handle(symbols)
	err = jsonHelper.WriteJSONFile(sourceProvider.GetArbitragePairCachePath(), triangularPairs)
	helpers.Panic(err)
}
