package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/models"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceprovider"
	"arbitrage-bot/sourceprovider/dex"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"time"
)

const TriangularPairLimit = 500

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
	currentSourceProviderName := sourceprovider.SourceProviderName["Uniswap"]
	sourceProvider := dex.GetSourceProvider(currentSourceProviderName)
	arbitrageCalculator := arbitrage.NewAmmArbitrageCalculator(sourceProvider)
	//currentSourceProviderName := sourceprovider.SourceProviderName["Binance"]
	//sourceProvider := cex.GetSourceProvider(currentSourceProviderName)
	//arbitrageCalculator := arbitrage.NewArbitrageCalculator(sourceProvider)

	var triangularPairBatches = step1(sourceProvider, false)
	_ = triangularPairBatches
	var symbols []*sourceprovider.Symbol

	if len(triangularPairBatches) > TriangularPairLimit {
		triangularPairBatches = triangularPairBatches[:TriangularPairLimit]
	}

	for _, pair := range triangularPairBatches {
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
		var surfaceResults []models.TriangularArbSurfaceResult

		for _, triangularPairs := range triangularPairBatches {
			var startingAmount float64 = 5
			surfaceResult, _ := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs, startingAmount)

			if surfaceResult.ProfitLoss > 0 {
				surfaceResults = append(surfaceResults, surfaceResult)
				//} else {
				//fmt.Println(err)
			}
		}

		// TODO: return stream of results
		fmt.Println("Fetching depth for the surface results...")
		results, err := arbitrageCalculator.BatchGetDepth(surfaceResults)
		helpers.Panic(err)
		fmt.Println("results", results)

		for _, result := range results {
			if result[0].ProfitLoss > 0 || result[1].ProfitLoss > 0 {
				fmt.Println(result)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
