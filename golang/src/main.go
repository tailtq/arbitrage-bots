package main

import (
	"arbitrage-bot/helpers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/models"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"arbitrage-bot/services/web3"
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
	//sourceProvider := dex.NewUniswapSourceProviderService()
	sourceProvider := dex.NewPancakeswapSourceProvider()
	arbitrageCalculator := arbitrage.NewAmmArbitrageCalculator(sourceProvider)
	arbitrageExecutor := web3.NewArbitrageExecutorWeb3Service()

	// for networks like base, celo, we'll run a command to obtain the triangular pairs, then get cache from step1
	var triangularPairBatches = step1(sourceProvider)
	var symbols []*sourceprovider.Symbol

	for _, pair := range triangularPairBatches {
		for _, symbol := range pair {
			symbols = append(symbols, symbol)
		}
	}
	var pingChannel = make(chan bool)
	var startingAmount float64 = 5
	var verbose = false
	go sourceProvider.SubscribeSymbols(symbols, pingChannel, verbose)

	fmt.Println("Subscribed to symbols, waiting for data...")
	time.Sleep(3 * time.Second)
	fmt.Println("Starting the arbitrage calculation...")

	for range pingChannel {
		var surfaceResults []models.TriangularArbSurfaceResult

		for _, triangularPairs := range triangularPairBatches {
			surfaceResult, _ := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs, startingAmount)

			if surfaceResult.ProfitLoss > 0 {
				surfaceResults = append(surfaceResults, surfaceResult)
			}
		}

		if len(surfaceResults) > 0 {
			fmt.Println("Fetching depth for the surface results...")

			for _, surfaceRate := range surfaceResults {
				var depthResult = arbitrageCalculator.CalcDepthOpportunityForward(surfaceRate, verbose)

				// execute the arbitrage if the profit is between 1% and 10%
				if depthResult.ProfitLossPerc > 0.01 && depthResult.ProfitLossPerc < 0.1 {
					var loanAddress = arbitrageExecutor.GetLoanAddress(symbols, depthResult.TradePaths)
					err := arbitrageExecutor.ExecuteArbitrage(depthResult.TradePaths, startingAmount, loanAddress)

					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("HAHAHAHA", depthResult)
				}
			}
		}

		fmt.Println("========================")
		time.Sleep(10 * time.Second)
	}
}
