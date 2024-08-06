package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceProvider"
	CEX "arbitrage-bot/sourceProvider/cex"
	"fmt"
	"time"
)

func step1(cexSourceProvider sourceProvider.SourceProviderInterface, force bool) [][3]*sourceProvider.Symbol {
	// get all the triangular pairs
	arbitragePairPath := cexSourceProvider.GetArbitragePairCachePath()

	if !force && fileHelper.PathExists(arbitragePairPath) {
		var symbols [][3]*sourceProvider.Symbol
		err := jsonHelper.ReadJSONFile(arbitragePairPath, &symbols)
		helpers.Panic(err)

		return symbols
	}

	// NOTE: this doesn't cover the case when we have multiple CEX
	triangularPairFinder := arbitrage.TriangularPairFinder{}
	symbols, err := cexSourceProvider.GetSymbols(force)
	helpers.Panic(err)

	// find the arbitrage pairs -> cache it
	triangularPairs := triangularPairFinder.Handle(symbols, len(symbols))
	err = jsonHelper.WriteJSONFile(arbitragePairPath, triangularPairs)
	helpers.Panic(err)

	return triangularPairs
}

func main() {
	cexSourceProvider := CEX.NewMEXCSourceProvider()
	var triangularPairBatches [][3]*sourceProvider.Symbol = step1(cexSourceProvider, false)
	var symbols []*sourceProvider.Symbol

	for _, pair := range triangularPairBatches {
		for _, symbol := range pair {
			symbols = append(symbols, symbol)
		}
	}

	cexSourceProvider.SubscribeSymbols(symbols)
	arbitrageCalculator := arbitrage.NewArbitrageCalculator(cexSourceProvider)

	fmt.Println("Subscribed to symbols, waiting for data...")
	time.Sleep(3 * time.Second)
	fmt.Println("Starting the arbitrage calculation...")

	for {
		for _, triangularPairs := range triangularPairBatches {
			startingAmount := 10
			result := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs, float64(startingAmount))

			if result != nil && result.ProfitLoss > arbitrage.MinSurfaceRate {
				depthResult := arbitrageCalculator.GetDepthFromOrderBook(result)
				fmt.Println(result)
				fmt.Println(depthResult)
				fmt.Println("---------")
			}
		}
		time.Sleep(3 * time.Second)
		fmt.Println("------")
	}
}
