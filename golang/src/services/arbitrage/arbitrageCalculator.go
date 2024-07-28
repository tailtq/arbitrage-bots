package arbitrage

import (
	"arbitrage-bot/sourceProvider"
	"fmt"
)

type ArbitrageCalculator struct {
	sourceProvider sourceProvider.SourceProvider
}

func NewArbitrageCalculator(sourceProvider sourceProvider.SourceProvider) *ArbitrageCalculator {
	return &ArbitrageCalculator{sourceProvider: sourceProvider}
}

func (a *ArbitrageCalculator) GetPriceForTriangularPair(triangularPair [3]*sourceProvider.Symbol) (*TriangularBidAskPrice, error) {
	// get the price data for the triangular pair
	symbolData := a.sourceProvider.SymbolPriceData()
	symbol1Price, ok1 := symbolData[triangularPair[0].Symbol]
	symbol2Price, ok2 := symbolData[triangularPair[1].Symbol]
	symbol3Price, ok3 := symbolData[triangularPair[2].Symbol]

	if !ok1 {
		err := fmt.Errorf("symbol %s not found", triangularPair[0].Symbol)
		return nil, err
	} else if !ok2 {
		err := fmt.Errorf("symbol %s not found", triangularPair[1].Symbol)
		return nil, err
	} else if !ok3 {
		err := fmt.Errorf("symbol %s not found", triangularPair[2].Symbol)
		return nil, err
	}


	return &TriangularBidAskPrice{
		pairAAsk: symbol1Price.BestAsk,
		pairABid: symbol1Price.BestBid,
		pairBAsk: symbol2Price.BestAsk,
		pairBBid: symbol2Price.BestBid,
		pairCAsk: symbol3Price.BestAsk,
		pairCBid: symbol3Price.BestBid,
	}, nil
}

func (a *ArbitrageCalculator) CalcTriangularArbSurfaceRate(triangularPair [3]*sourceProvider.Symbol) *TriangularSurfaceRate {
	priceData, err := a.GetPriceForTriangularPair(triangularPair)

	if err != nil {
		return nil
	}

	// set variables
	var startingAmount float64 = 1  // the amount of the asset that is used to calculate the arbitrage
	var contract1 string = ""
	var contract2 string = ""
	var contract3 string = ""
	var directionTrade1 string = ""
	var directionTrade2 string = ""
	var directionTrade3 string = ""
	var acquiredCoinT1 float64 = 0
	var acquiredCoinT2 float64 = 0
	var acquiredCoinT3 float64 = 0
	var calculated bool = false

	var aPair string = triangularPair[0].Symbol
	var aBase string = triangularPair[0].BaseAsset
	var aQuote string = triangularPair[0].QuoteAsset
	var bPair string = triangularPair[1].Symbol
	var bBase string = triangularPair[1].BaseAsset
	var bQuote string = triangularPair[1].QuoteAsset
	var cPair string = triangularPair[2].Symbol
	var cBase string = triangularPair[2].BaseAsset
	var cQuote string = triangularPair[2].QuoteAsset

	// set directions and loop through
	var directions = [2]string{"forward", "backward"}

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
			swap1Rate = 1 / priceData.pairAAsk
			directionTrade1 = "baseToQuote"
		}
		
		// Assume starting with aBase and swapping for aQuote
		if direction == "backward" {
			swap1 = aQuote
			swap2 = aBase
			swap1Rate = priceData.pairABid
			directionTrade1 = "quoteToBase"
		}
		// Place first trade
		contract1 = aPair
		acquiredCoinT1 = startingAmount * swap1Rate

		// FORWARD
		if direction == "forward" {
			// SCENARIO 1: aQuote matches bQuote
			if aQuote == bQuote && !calculated {
				// f.e. USDT_BTC -> ETH_BTC (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairBBid
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = bPair

				// If bBase (acquired coin 2) matches cBase
				if bBase == cBase {
					swap3 = cBase
					swap3Rate = 1 / priceData.pairCAsk
					directionTrade3 = "baseToQuote"
				}

				// If bBase (acquired coin 2) matches cQuote
				if bBase == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				calculated = true
			}

			// SCENARIO 2: aQuote matches bBase
			if aQuote == bBase && !calculated {
				// f.e. USDT_BTC -> BTC_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = 1 / priceData.pairBAsk
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = bPair

				// If bQuote (acquired coin 2) matches cBase
				if bQuote == cBase {
					swap3 = cBase
					swap3Rate = 1 / priceData.pairCAsk
					directionTrade3 = "baseToQuote"
				}

				// If bQuote (acquired coin 2) matches cQuote
				if bQuote == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				calculated = true
			}

			// SCENARIO 3: aQuote matches cQuote
			if aQuote == cQuote && !calculated {
				// f.e. USDT_BTC -> ETH_BTC (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairCBid
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = cPair

				// If cBase (acquired coin 2) matches bBase
				if cBase == bBase {
					swap3 = bBase
					swap3Rate = 1 / priceData.pairBAsk
					directionTrade3 = "baseToQuote"
				}

				// If cBase (acquired coin 2) matches bQuote
				if cBase == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				calculated = true
			}

			// SCENARIO 4: aQuote matches cBase
			if aQuote == cBase && !calculated {
				// f.e. USDT_BTC -> BTC_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = 1 / priceData.pairCAsk
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = cPair

				// If cQuote (acquired coin 2) matches bBase
				if cQuote == bBase {
					swap3 = bBase
					swap3Rate = 1 / priceData.pairBAsk
					directionTrade3 = "baseToQuote"
				}

				// If cQuote (acquired coin 2) matches bQuote
				if cQuote == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				calculated = true
			}
		}
		
		// BACKWARD
		if direction == "backward" {
			// SCENARIO 1: aBase matches bQuote
			if aBase == bQuote && !calculated {
				// f.e. USDT_BTC -> ETH_USDT (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairBBid
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = bPair

				// If bBase (acquired coin 2) matches cBase
				if bBase == cBase {
					swap3 = cBase
					swap3Rate = 1 / priceData.pairCAsk
					directionTrade3 = "baseToQuote"
				}

				// If bBase (acquired coin 2) matches cQuote
				if bBase == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				calculated = true
			}

			// SCENARIO 2: aBase matches bBase
			if aBase == bBase && !calculated {
				// f.e. USDT_BTC -> USDT_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = 1 / priceData.pairBAsk
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = bPair

				// If bQuote (acquired coin 2) matches cBase
				if bQuote == cBase {
					swap3 = cBase
					swap3Rate = 1 / priceData.pairCAsk
					directionTrade3 = "baseToQuote"
				}

				// If bQuote (acquired coin 2) matches cQuote
				if bQuote == cQuote {
					swap3 = cQuote
					swap3Rate = priceData.pairCBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = cPair
				calculated = true
			}

			// SCENARIO 3: aBase matches cQuote
			if aBase == cQuote && !calculated {
				// f.e. USDT_BTC -> ETH_USDT (acquiredCoinT2 would be ETH, it's backward so we'll use the bid price)
				swap2Rate = priceData.pairCBid
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "quoteToBase"
				contract2 = cPair

				// If cBase (acquired coin 2) matches bBase
				if cBase == bBase {
					swap3 = bBase
					swap3Rate = 1 / priceData.pairBAsk
					directionTrade3 = "baseToQuote"
				}

				// If cBase (acquired coin 2) matches bQuote
				if cBase == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				calculated = true
			}

			// SCENARIO 4: aBase matches cBase
			if aBase == cBase && !calculated {
				// f.e. USDT_BTC -> USDT_ETH (acquiredCoinT2 would be ETH, it's forward so we'll use the ask price)
				swap2Rate = 1 / priceData.pairCAsk
				acquiredCoinT2 = acquiredCoinT1 * swap2Rate
				directionTrade2 = "baseToQuote"
				contract2 = cPair

				// If cQuote (acquired coin 2) matches bBase
				if cQuote == bBase {
					swap3 = bBase
					swap3Rate = 1 / priceData.pairBAsk
					directionTrade3 = "baseToQuote"
				}

				// If cQuote (acquired coin 2) matches bQuote
				if cQuote == bQuote {
					swap3 = bQuote
					swap3Rate = priceData.pairBBid
					directionTrade3 = "quoteToBase"
				}

				acquiredCoinT3 = acquiredCoinT2 * swap3Rate
				contract3 = bPair
				calculated = true
			}
		}

		// PROFIT LOSS OUTPUT
		// Profit and loss calculation
		var profitLoss float64 = acquiredCoinT3 - startingAmount
		var profitLossPercentage float32 = float32(profitLoss / startingAmount) * 100

		// Trade Descriptions
		var tradeDescription1 string = fmt.Sprintf("Start with %v of %v, swap at %v for %v, acquiring %v", swap1, startingAmount, swap1Rate, swap2, acquiredCoinT1)
		var tradeDescription2 string = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT1, swap2, swap2Rate, swap3, acquiredCoinT2)
		var tradeDescription3 string = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT2, swap3, swap3Rate, swap1, acquiredCoinT3)

		if profitLoss > MinSurfaceRate {
			return &TriangularSurfaceRate{
				Swap1:             swap1,
				Swap2:             swap2,
				Swap3:             swap3,
				Contract1:         contract1,
				Contract2:         contract2,
				Contract3:         contract3,
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
		}
	}

	return nil
}
