package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceprovider"
	"arbitrage-bot/sourceprovider/dex"
	"fmt"
	"time"
)

func step1(sourceProvider sourceprovider.ISourceProvider, force bool) [][3]*sourceprovider.Symbol {
	// get all the triangular pairs
	arbitragePairPath := sourceProvider.GetArbitragePairCachePath()

	if !force && fileHelper.PathExists(arbitragePairPath) {
		var symbols [][3]*sourceprovider.Symbol
		err := jsonHelper.ReadJSONFile(arbitragePairPath, &symbols)
		helpers.Panic(err)

		return symbols
	}

	// NOTE: this doesn't cover the case when we have multiple CEX
	triangularPairFinder := arbitrage.TriangularPairFinder{}
	symbols, err := sourceProvider.GetSymbols(force)
	helpers.Panic(err)

	// find the arbitrage pairs -> cache it
	triangularPairs := triangularPairFinder.Handle(symbols, len(symbols))
	err = jsonHelper.WriteJSONFile(arbitragePairPath, triangularPairs)
	helpers.Panic(err)

	return triangularPairs
}

// CEX/DEX arbitrage opportunities
func main() {
	const TriangularPairLimit = 500
	currentSourceProviderName := sourceprovider.SourceProviderName["Uniswap"]
	sourceProvider := dex.GetSourceProvider(currentSourceProviderName)
	arbitrageCalculator := arbitrage.NewAmmArbitrageCalculator(sourceProvider)
	//currentSourceProviderName := sourceprovider.SourceProviderName["Binance"]
	//sourceProvider := cex.GetSourceProvider(currentSourceProviderName)
	//arbitrageCalculator := arbitrage.NewArbitrageCalculator(sourceProvider)

	var triangularPairBatches = step1(sourceProvider, false)
	_ = triangularPairBatches
	var symbols []*sourceprovider.Symbol

	for _, pair := range triangularPairBatches[:TriangularPairLimit] {
		for _, symbol := range pair {
			symbols = append(symbols, symbol)
		}
	}

	if currentSourceProviderName == sourceprovider.SourceProviderName["Uniswap"] {
		go sourceProvider.SubscribeSymbols(symbols)
	} else {
		sourceProvider.SubscribeSymbols(symbols)
	}

	fmt.Println("Subscribed to symbols, waiting for data...")
	time.Sleep(3 * time.Second)
	fmt.Println("Starting the arbitrage calculation...")

	for {
		for _, triangularPairs := range triangularPairBatches {
			startingAmount := 10
			result, err := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs, float64(startingAmount))

			if err == nil {
				fmt.Println(result.Contract1, result.Contract2, result.Contract3, result.ProfitLossPerc, result.ProfitLoss)
				//depthResult, err2 := arbitrageCalculator.GetDepthFromOrderBook(result)

				//if err2 == nil {
				//	fmt.Println(result)
				//	fmt.Println(depthResult)
				//	fmt.Println("---------")
				//}
			}
		}
		fmt.Println("------")
		time.Sleep(3 * time.Second)
	}
}
