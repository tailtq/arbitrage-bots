package arbitrage

import (
	"arbitrage-bot/sourceProvider"
	"fmt"
)

// ArbitrageCalculator ... the main calculator for the arbitrage
type ArbitrageCalculator struct {
	sourceProvider sourceProvider.ISourceProvider
}

// NewArbitrageCalculator ... creates a new instance of the ArbitrageCalculator
func NewArbitrageCalculator(sourceProvider sourceProvider.ISourceProvider) *ArbitrageCalculator {
	return &ArbitrageCalculator{sourceProvider: sourceProvider}
}

// GetPriceForTriangularPair ... get the price data for the triangular pair
func (a *ArbitrageCalculator) GetPriceForTriangularPair(triangularPair [3]*sourceProvider.Symbol) (*TriangularBidAskPrice, error) {
	symbol1Price := a.sourceProvider.GetSymbolPrice(triangularPair[0].Symbol)
	symbol2Price := a.sourceProvider.GetSymbolPrice(triangularPair[1].Symbol)
	symbol3Price := a.sourceProvider.GetSymbolPrice(triangularPair[2].Symbol)

	if symbol1Price == nil {
		err := fmt.Errorf("symbol %s not found", triangularPair[0].Symbol)
		return nil, err
	} else if symbol2Price == nil {
		err := fmt.Errorf("symbol %s not found", triangularPair[1].Symbol)
		return nil, err
	} else if symbol3Price == nil {
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

// CalcTriangularArbSurfaceRate ... calculates the surface rate for the triangular pair.
func (a *ArbitrageCalculator) CalcTriangularArbSurfaceRate(triangularPair [3]*sourceProvider.Symbol, startingAmount float64) *TriangularSurfaceTradingResult {
	priceData, err := a.GetPriceForTriangularPair(triangularPair)

	if err != nil {
		return nil
	}

	// set variables
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
		var profitLossPercentage float32 = float32(profitLoss/startingAmount) * 100

		// Trade Descriptions
		var tradeDescription1 string = fmt.Sprintf("Start with %v of %v, swap at %v for %v, acquiring %v", swap1, startingAmount, swap1Rate, swap2, acquiredCoinT1)
		var tradeDescription2 string = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT1, swap2, swap2Rate, swap3, acquiredCoinT2)
		var tradeDescription3 string = fmt.Sprintf("Swap %v of %v at %v for %v, acquiring %v", acquiredCoinT2, swap3, swap3Rate, swap1, acquiredCoinT3)

		return &TriangularSurfaceTradingResult{
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

	return nil
}

// reformatOrderbook ... reformat the orderbook to be used in the calculation
func (a *ArbitrageCalculator) reformatOrderbook(
	directionTrade string, orderBookPrice *sourceProvider.SymbolOrderbookDepth,
) []*sourceProvider.OrderbookEntry {
	var result []*sourceProvider.OrderbookEntry

	if directionTrade == "baseToQuote" {
		for _, entry := range orderBookPrice.Asks {
			// TODO: add comment to this
			var adjPrice float64 = 0

			if entry.Price != 0 {
				adjPrice = 1 / entry.Price
			}

			var adjQuantity float64 = entry.Quantity * adjPrice
			result = append(result, &sourceProvider.OrderbookEntry{
				Price:    adjPrice,
				Quantity: adjQuantity,
			})
		}
	} else if directionTrade == "quoteToBase" {
		for _, entry := range orderBookPrice.Bids {
			var adjPrice float64 = entry.Price
			var adjQuantity float64 = entry.Quantity
			result = append(result, &sourceProvider.OrderbookEntry{
				Price:    adjPrice,
				Quantity: adjQuantity,
			})
		}
	}

	return result
}

// CalculateAcquiredCoin ... get acquired coin also known as (aka) Depth calculation
func (a *ArbitrageCalculator) calculateAcquiredCoin(amountIn float64, orderbook []*sourceProvider.OrderbookEntry) float64 {
	// CHALLENGES:
	// - Full amount of starting amount in can be eaten on the first level (level 0)
	// - Some of the amount in can be eaten up by multiple levels
	// - Some coins may not have enough liquidity

	tradingBalance := amountIn
	quantityBought := 0.0
	acquiredCoin := 0.0
	counts := 0

	for _, level := range orderbook {
		var amountBought float64 = 0

		// Amount In <= first level's total amount
		if level.Quantity >= tradingBalance {
			quantityBought = tradingBalance
			tradingBalance = 0
			amountBought = quantityBought * level.Price
		}
		// Amount In > a given level's total amount
		if level.Quantity < tradingBalance {
			quantityBought = level.Quantity
			tradingBalance -= level.Quantity
			amountBought = quantityBought * level.Price
		}

		acquiredCoin += amountBought

		if tradingBalance == 0 {
			return acquiredCoin
		}

		counts++

		if counts == len(orderbook) {
			return 0
		}
	}
	return 0
}

func (a *ArbitrageCalculator) GetDepthFromOrderBook(surfaceRate *TriangularSurfaceTradingResult) *TriangularDepthTradingResult {
	var startingAmount float64 = surfaceRate.StartingAmount

	// Define variables
	var contract1 string = surfaceRate.Contract1
	var contract2 string = surfaceRate.Contract2
	var contract3 string = surfaceRate.Contract3
	var directionTrade1 string = surfaceRate.DirectionTrade1
	var directionTrade2 string = surfaceRate.DirectionTrade2
	var directionTrade3 string = surfaceRate.DirectionTrade3
	var depthContract1 *sourceProvider.SymbolOrderbookDepth = a.sourceProvider.GetSymbolOrderbookDepth(contract1)
	var depthContract2 *sourceProvider.SymbolOrderbookDepth = a.sourceProvider.GetSymbolOrderbookDepth(contract2)
	var depthContract3 *sourceProvider.SymbolOrderbookDepth = a.sourceProvider.GetSymbolOrderbookDepth(contract3)

	if depthContract1 == nil {
		fmt.Printf("Error: depthContract1 %v is nil\n", contract1)
		return nil
	} else if depthContract2 == nil {
		fmt.Printf("Error: depthContract2 %v is nil\n", contract2)
		return nil
	} else if depthContract3 == nil {
		fmt.Printf("Error: depthContract3 %v is nil\n", contract3)
		return nil
	}

	// get acquired coins
	orderbook1 := a.reformatOrderbook(directionTrade1, depthContract1)
	orderbook2 := a.reformatOrderbook(directionTrade2, depthContract2)
	orderbook3 := a.reformatOrderbook(directionTrade3, depthContract3)
	acquiredCoinT1 := a.calculateAcquiredCoin(startingAmount, orderbook1)
	acquiredCoinT2 := a.calculateAcquiredCoin(acquiredCoinT1, orderbook2)
	acquiredCoinT3 := a.calculateAcquiredCoin(acquiredCoinT2, orderbook3)

	// calculate profit loss also known as real rate
	profitLoss := acquiredCoinT3 - startingAmount
	realRatePercent := 0.0

	if profitLoss != 0 {
		realRatePercent = (profitLoss / startingAmount) * 100
	}

	if realRatePercent > -1 {
		return &TriangularDepthTradingResult{
			ProfitLoss:     profitLoss,
			ProfitLossPerc: float32(realRatePercent),
		}
	}
	return nil
}
