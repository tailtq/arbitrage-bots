package commands

import (
	"arbitrage-bot/helpers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"arbitrage-bot/services/web3"
	"fmt"
	"sync"
)

type FetchPancakeswapPoolDataCommand struct {
	web3Service *web3.PancakeswapWeb3Service
}

// NewFetchPancakeswapPoolDataCommand ... creates a new FetchPancakeswapPoolDataCommand
func NewFetchPancakeswapPoolDataCommand() *FetchPancakeswapPoolDataCommand {
	return &FetchPancakeswapPoolDataCommand{
		web3Service: web3.NewPancakeswapWeb3Service(),
	}
}

// fetchSymbolsFromNetwork ... fetches symbols from the network
func (c *FetchPancakeswapPoolDataCommand) fetchSymbols() []*sourceprovider.Symbol {
	var symbols []*sourceprovider.Symbol
	var maxPools = 1000
	var concurrency = 5
	var channel = make(chan int)
	var wg = sync.WaitGroup{}
	wg.Add(concurrency)

	// Fetch pool data concurrently (5 coroutines)
	for range concurrency {
		go func() {
			defer wg.Done()
			for index := range channel {
				var symbol = c.web3Service.GetPoolDataByIndex(index)
				symbols = append(symbols, &symbol)
			}
		}()
	}
	for index := range maxPools {
		channel <- index
	}
	close(channel)
	wg.Wait()

	return symbols
}

// Fetch ... fetches Pancake pool data & find triangular pairs
func (c *FetchPancakeswapPoolDataCommand) Fetch() {
	// Fetch symbols from the network
	var symbols = c.fetchSymbols()
	fmt.Println("Fetched", len(symbols), "symbols")

	// Find triangular pairs & save to cache
	var sourceProvider = dex.NewPancakeswapSourceProvider()
	var triangularPairFinder = arbitrage.TriangularPairFinder{}
	triangularPairs := triangularPairFinder.Handle(symbols)
	var err = jsonHelper.WriteJSONFile(sourceProvider.GetArbitragePairCachePath(), triangularPairs)
	helpers.Panic(err)
}
