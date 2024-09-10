package arbitrage

import (
	"arbitrage-bot/models"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"fmt"
)

// AmmArbitrageCalculator ... the main AMM calculator for the arbitrage (automated market maker)
type AmmArbitrageCalculator struct {
	sourceProvider dex.ISourceProvider
}

// NewAmmArbitrageCalculator ... creates a new instance of the AmmArbitrageCalculator
func NewAmmArbitrageCalculator(sourceProvider dex.ISourceProvider) *AmmArbitrageCalculator {
	return &AmmArbitrageCalculator{sourceProvider: sourceProvider}
}

// GetPriceForTriangularPair ... get the price data for the triangular pair
func (a *AmmArbitrageCalculator) getPriceForTriangularPair(triangularPair [3]*sourceprovider.Symbol) (TriangularDexPrice, error) {
	symbol1Price := a.sourceProvider.GetSymbolPrice(triangularPair[0].Symbol)
	symbol2Price := a.sourceProvider.GetSymbolPrice(triangularPair[1].Symbol)
	symbol3Price := a.sourceProvider.GetSymbolPrice(triangularPair[2].Symbol)

	if symbol1Price == nil {
		err := fmt.Errorf("symbol %s not found", triangularPair[0].Symbol)
		return TriangularDexPrice{}, err
	} else if symbol2Price == nil {
		err := fmt.Errorf("symbol %s not found", triangularPair[1].Symbol)
		return TriangularDexPrice{}, err
	} else if symbol3Price == nil {
		err := fmt.Errorf("symbol %s not found", triangularPair[2].Symbol)
		return TriangularDexPrice{}, err
	}

	return TriangularDexPrice{
		pairAToken0: symbol1Price.Token0Price,
		pairAToken1: symbol1Price.Token1Price,
		pairBToken0: symbol2Price.Token0Price,
		pairBToken1: symbol2Price.Token1Price,
		pairCToken0: symbol3Price.Token0Price,
		pairCToken1: symbol3Price.Token1Price,
	}, nil
}

// CalcTriangularArbSurfaceRate ... calculates the surface rate for the triangular pair.
func (a *AmmArbitrageCalculator) CalcTriangularArbSurfaceRate(triangularPair [3]*sourceprovider.Symbol, startingAmount float64) (models.TriangularArbSurfaceResult, error) {
	priceData, err := a.getPriceForTriangularPair(triangularPair)

	if err != nil {
		return models.TriangularArbSurfaceResult{}, err
	}

	// set variables
	var contract1 = ""
	var contract2 = ""
	var contract3 = ""
	var contract1Address = ""
	var contract2Address = ""
	var contract3Address = ""
	var directionTrade1 = ""
	var directionTrade2 = ""
	var directionTrade3 = ""
	var acquiredCoinT1 float64 = 0
	var acquiredCoinT2 float64 = 0
	var acquiredCoinT3 float64 = 0
	var calculated = false

	var aPair = triangularPair[0].Symbol
	var aPairContractAddress = triangularPair[0].Address
	var aBase = triangularPair[0].BaseAsset
	var aQuote = triangularPair[0].QuoteAsset

	var bPair = triangularPair[1].Symbol
	var bPairContractAddress = triangularPair[1].Address
	var bBase = triangularPair[1].BaseAsset
	var bQuote = triangularPair[1].QuoteAsset

	var cPair = triangularPair[2].Symbol
	var cPairContractAddress = triangularPair[2].Address
	var cBase = triangularPair[2].BaseAsset
	var cQuote = triangularPair[2].QuoteAsset
	// set directions and loop through
	var directions = [2]string{"forward", "backward"}
	var tradingResult models.TriangularArbSurfaceResult

	for _, direction := range directions {
		// set additional variables for swap information
		var swap1 string
		var swap2 string
		var swap3 string
		var swap1Rate float64 = 0
		var swap2Rate float64 = 0
		var swap3Rate float64 = 0

		// If we are swapping the coin on the left (Base) to the right (Quote) then * (1/ Ask)
		// If we are swapping the coin on the right (Quote) to the left (Base) then * Bid

		// Assume starting with aBase and swapping for aQuote
		if direction == "forward" {
			swap1 = aBase
			swap2 = aQuote
			swap1Rate = priceData.pairAToken1
			directionTrade1 = "baseToQuote"
		}

		// Assume starting with aBase and swapping for aQuote
		if direction == "backward" {
			swap1 = aQuote
			swap2 = aBase
			swap1Rate = priceData.pairAToken0
			directionTrade1 = "quoteToBase"
		}
		// Place first trade
		contract1 = aPair
		contract1Address = aPairContractAddress
		acquiredCoinT1 = startingAmount * swap1Rate

		// FORWARD
		if direction == "forward" {
			// SCENARIO 1: aQuote matches bQuote
			if aQuote == bQuote && !calculated {
				// f.e. USDT_BTC -> ETH_BTC (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairBToken0
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = bPair
				contract2Address = bPairContractAddress

				// If bBase (acquired coin 2) matches cBase
				if bBase == cBase {
					swap3 = cBase
					swap3Rate = priceData.pairCToken1
					directionTrade3 = "baseToQuote"
				}

				// If bBase (acquired coin 2) matches cQuote
				if bBase == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				contract3Address = cPairContractAddress
				calculated = true
			}

			// SCENARIO 2: aQuote matches bBase
			if aQuote == bBase && !calculated {
				// f.e. USDT_BTC -> BTC_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = priceData.pairBToken1
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = bPair
				contract2Address = bPairContractAddress

				// If bQuote (acquired coin 2) matches cBase
				if bQuote == cBase {
					swap3 = cBase
					swap3Rate = priceData.pairCToken1
					directionTrade3 = "baseToQuote"
				}

				// If bQuote (acquired coin 2) matches cQuote
				if bQuote == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				contract3Address = cPairContractAddress
				calculated = true
			}

			// SCENARIO 3: aQuote matches cQuote
			if aQuote == cQuote && !calculated {
				// f.e. USDT_BTC -> ETH_BTC (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairCToken0
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = cPair
				contract2Address = cPairContractAddress

				// If cBase (acquired coin 2) matches bBase
				if cBase == bBase {
					swap3 = bBase
					swap3Rate = priceData.pairBToken1
					directionTrade3 = "baseToQuote"
				}

				// If cBase (acquired coin 2) matches bQuote
				if cBase == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				contract3Address = bPairContractAddress
				calculated = true
			}

			// SCENARIO 4: aQuote matches cBase
			if aQuote == cBase && !calculated {
				// f.e. USDT_BTC -> BTC_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = priceData.pairCToken1
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = cPair
				contract2Address = cPairContractAddress

				// If cQuote (acquired coin 2) matches bBase
				if cQuote == bBase {
					swap3 = bBase
					swap3Rate = priceData.pairBToken1
					directionTrade3 = "baseToQuote"
				}

				// If cQuote (acquired coin 2) matches bQuote
				if cQuote == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				contract3Address = bPairContractAddress
				calculated = true
			}
		}

		// BACKWARD
		if direction == "backward" {
			// SCENARIO 1: aBase matches bQuote
			if aBase == bQuote && !calculated {
				// f.e. USDT_BTC -> ETH_USDT (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairBToken0
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = bPair
				contract2Address = bPairContractAddress

				// If bBase (acquired coin 2) matches cBase
				if bBase == cBase {
					swap3 = cBase
					swap3Rate = priceData.pairCToken1
					directionTrade3 = "baseToQuote"
				}

				// If bBase (acquired coin 2) matches cQuote
				if bBase == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				contract3Address = cPairContractAddress
				calculated = true
			}

			// SCENARIO 2: aBase matches bBase
			if aBase == bBase && !calculated {
				// f.e. USDT_BTC -> USDT_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = priceData.pairBToken1
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = bPair
				contract2Address = bPairContractAddress

				// If bQuote (acquired coin 2) matches cBase
				if bQuote == cBase {
					swap3 = cBase
					swap3Rate = priceData.pairCToken1
					directionTrade3 = "baseToQuote"
				}

				// If bQuote (acquired coin 2) matches cQuote
				if bQuote == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				contract3Address = cPairContractAddress
				calculated = true
			}

			// SCENARIO 3: aBase matches cQuote
			if aBase == cQuote && !calculated {
				// f.e. USDT_BTC -> ETH_USDT (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairCToken0
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = cPair
				contract2Address = cPairContractAddress

				// If cBase (acquired coin 2) matches bBase
				if cBase == bBase {
					swap3 = bBase
					swap3Rate = priceData.pairBToken1
					directionTrade3 = "baseToQuote"
				}

				// If cBase (acquired coin 2) matches bQuote
				if cBase == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				contract3Address = bPairContractAddress
				calculated = true
			}

			// SCENARIO 4: aBase matches cBase
			if aBase == cBase && !calculated {
				// f.e. USDT_BTC -> USDT_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = priceData.pairCToken1
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = cPair
				contract2Address = cPairContractAddress

				// If cQuote (acquired coin 2) matches bBase
				if cQuote == bBase {
					swap3 = bBase
					swap3Rate = priceData.pairBToken1
					directionTrade3 = "baseToQuote"
				}

				// If cQuote (acquired coin 2) matches bQuote
				if cQuote == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBToken0
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				contract3Address = bPairContractAddress
				calculated = true
			}
		}

		// PROFIT LOSS OUTPUT
		// Profit and loss calculation
		var profitLoss = acquiredCoinT3 - startingAmount
		var profitLossPercentage = profitLoss / startingAmount * 100

		// Trade Descriptions
		var tradeDescription1 = fmt.Sprintf("Start with %v of %v, swap at %v for %v, acquiring %v", swap1, startingAmount, swap1Rate, swap2, acquiredCoinT1)
		var tradeDescription2 = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT1, swap2, swap2Rate, swap3, acquiredCoinT2)
		var tradeDescription3 = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT2, swap3, swap3Rate, swap1, acquiredCoinT3)
		tradingResult = models.TriangularArbSurfaceResult{
			Swap1:             swap1,
			Swap2:             swap2,
			Swap3:             swap3,
			Contract1:         contract1,
			Contract2:         contract2,
			Contract3:         contract3,
			Symbol1:           a.sourceProvider.GetSymbol(contract1),
			Symbol2:           a.sourceProvider.GetSymbol(contract2),
			Symbol3:           a.sourceProvider.GetSymbol(contract3),
			Contract1Address:  contract1Address,
			Contract2Address:  contract2Address,
			Contract3Address:  contract3Address,
			DirectionTrade1:   directionTrade1,
			DirectionTrade2:   directionTrade2,
			DirectionTrade3:   directionTrade3,
			StartingAmount:    startingAmount,
			AcquiredCoinT1:    acquiredCoinT1,
			AcquiredCoinT2:    acquiredCoinT2,
			AcquiredCoinT3:    acquiredCoinT3,
			Swap1Rate:         swap1Rate,
			Swap2Rate:         swap2Rate,
			Swap3Rate:         swap3Rate,
			ProfitLoss:        profitLoss,
			ProfitLossPerc:    profitLossPercentage,
			Direction:         direction,
			TradeDescription1: tradeDescription1,
			TradeDescription2: tradeDescription2,
			TradeDescription3: tradeDescription3,
		}

		if profitLoss > MinSurfaceRate {
			return tradingResult, nil
		}
	}

	return tradingResult, fmt.Errorf("no profitable arbitrage found")
}

// BatchCalcDepth ... get the depth from DEX (uniswap, ...) for multiple surface rates
func (a *AmmArbitrageCalculator) BatchCalcDepth(surfaceRates []models.TriangularArbSurfaceResult) []models.TriangularArbFullResult {
	var results []models.TriangularArbFullResult

	for _, surfaceRate := range surfaceRates {
		results = append(results, models.TriangularArbFullResult{
			SurfaceResult:       surfaceRate,
			DepthResultForward:  a.calcDepthOpportunityForward(surfaceRate),
			DepthResultBackward: a.calcDepthOpportunityBackward(surfaceRate),
		})
	}

	return results
}

func (a *AmmArbitrageCalculator) calcDepthOpportunityForward(surfaceResult models.TriangularArbSurfaceResult) models.TriangularArbDepthResult {
	var contract1 = surfaceResult.Contract1
	var contract2 = surfaceResult.Contract2
	var contract3 = surfaceResult.Contract3
	var directionTrade1 = surfaceResult.DirectionTrade1
	var directionTrade2 = surfaceResult.DirectionTrade2
	var directionTrade3 = surfaceResult.DirectionTrade3
	var allContracts = fmt.Sprintf("%s_%s_%s", contract1, contract2, contract3)
	_ = allContracts

	var acquiredCoinT1 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol1, surfaceResult.StartingAmount, directionTrade1, true)
	var acquiredCoinT2 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol2, acquiredCoinT1, directionTrade2, true)
	var acquiredCoinT3 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol3, acquiredCoinT2, directionTrade3, true)

	return a.calcDepthArb(surfaceResult.StartingAmount, acquiredCoinT3)
}

func (a *AmmArbitrageCalculator) calcDepthOpportunityBackward(surfaceResult models.TriangularArbSurfaceResult) models.TriangularArbDepthResult {
	var contract1 = surfaceResult.Contract3
	var contract2 = surfaceResult.Contract2
	var contract3 = surfaceResult.Contract1
	var directionTrade1 = a.revertDirection(surfaceResult.DirectionTrade3)
	var directionTrade2 = a.revertDirection(surfaceResult.DirectionTrade2)
	var directionTrade3 = a.revertDirection(surfaceResult.DirectionTrade1)
	var allContracts = fmt.Sprintf("%s_%s_%s", contract1, contract2, contract3)
	_ = allContracts

	var acquiredCoinT1 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol3, surfaceResult.StartingAmount, directionTrade1, true)
	var acquiredCoinT2 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol2, acquiredCoinT1, directionTrade2, true)
	var acquiredCoinT3 = a.sourceProvider.Web3Service().GetPrice(surfaceResult.Symbol1, acquiredCoinT2, directionTrade3, true)

	return a.calcDepthArb(surfaceResult.StartingAmount, acquiredCoinT3)
}

func (a *AmmArbitrageCalculator) revertDirection(direction string) string {
	if direction == "baseToQuote" {
		return "quoteToBase"
	} else if direction == "quoteToBase" {
		return "baseToQuote"
	}
	return ""
}

// calcDepthArb ... calculate the depth arbitrage
func (a *AmmArbitrageCalculator) calcDepthArb(
	amountIn float64,
	outputOut float64,
) models.TriangularArbDepthResult {
	var profitLoss = outputOut - amountIn
	var profitLossPerc = (profitLoss / amountIn) * 100

	return models.TriangularArbDepthResult{
		ProfitLoss:     profitLoss,
		ProfitLossPerc: profitLossPerc,
	}
}
