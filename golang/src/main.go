package main

import (
	"arbitrage-bot/helpers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/models"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"time"
)

func step1(sourceProvider sourceprovider.ISourceProvider) [][3]*sourceprovider.Symbol {
	// get cached arbitrage pairs (need to run command to fetch if not exists)
	arbitragePairPath := sourceProvider.GetArbitragePairCachePath()
	var symbols [][3]*sourceprovider.Symbol
	err := jsonHelper.ReadJSONFile(arbitragePairPath, &symbols)
	helpers.Panic(err)

	return symbols

	//if !force && fileHelper.PathExists(arbitragePairPath) {
	//
	//}
	//
	//// NOTE: this doesn't cover the case when we have multiple CEX
	//triangularPairFinder := arbitrage.TriangularPairFinder{}
	//symbols, err := sourceProvider.GetSymbols(force)
	//helpers.Panic(err)
	//
	//// find the arbitrage pairs -> cache it
	//triangularPairs := triangularPairFinder.Handle(symbols)
	//err = jsonHelper.WriteJSONFile(arbitragePairPath, triangularPairs)
	//helpers.Panic(err)
	//
	//return triangularPairs
}

// CEX/DEX arbitrage opportunities
func main() {
	sourceProvider := dex.NewUniswapSourceProviderService()
	arbitrageCalculator := arbitrage.NewAmmArbitrageCalculator(sourceProvider)

	// for networks like base, celo, we'll run a command to obtain the triangular pairs, then get cache from step1
	var triangularPairBatches = step1(sourceProvider)
	var symbols []*sourceprovider.Symbol

	for _, pair := range triangularPairBatches {
		for _, symbol := range pair {
			symbols = append(symbols, symbol)
		}
	}

	var pingChannel = make(chan bool)
	go sourceProvider.SubscribeSymbols(symbols, pingChannel)

	fmt.Println("Subscribed to symbols, waiting for data...")
	time.Sleep(3 * time.Second)
	fmt.Println("Starting the arbitrage calculation...")

	for range pingChannel {
		var surfaceResults []models.TriangularArbSurfaceResult

		for _, triangularPairs := range triangularPairBatches {
			var startingAmount float64 = 5
			surfaceResult, _ := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs, startingAmount)

			if surfaceResult.ProfitLoss > 0 {
				surfaceResults = append(surfaceResults, surfaceResult)
			}
		}

		if len(surfaceResults) > 0 {
			fmt.Println("Fetching depth for the surface results...")
			var depthResults = arbitrageCalculator.BatchCalcDepth(surfaceResults)

			for _, depthResult := range depthResults {
				if depthResult.DepthResultForward.ProfitLoss > 0 || depthResult.DepthResultBackward.ProfitLoss > 0 {
					fmt.Println("HAHAHAHA", depthResult)
				}
			}
		}

		fmt.Println("========================")
		time.Sleep(10 * time.Second)
	}
}
