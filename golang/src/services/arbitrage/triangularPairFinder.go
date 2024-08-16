package arbitrage

import (
	"arbitrage-bot/sourceprovider"
	"slices"
	"sort"
	"strings"
)

type TriangularPairFinder struct{}

func (t *TriangularPairFinder) Handle(symbols []*sourceprovider.Symbol, amount int) [][3]*sourceprovider.Symbol {
	// find a list of 3 arbitrage pairs (f.e. SEIBNB BNBBTC SEIBTC)
	var triangularPairsList [][3]*sourceprovider.Symbol
	var removeDuplicatesMap = make(map[string]bool)
	var pairsList = symbols[:amount]

	// get pair A
	for _, pairA := range pairsList {
		var aPairBox = []string{pairA.BaseAsset, pairA.QuoteAsset}

		// get pair B
		for _, pairB := range pairsList {
			if pairA == pairB {
				continue
			}

			// if three pairs form a cycle, continue
			if slices.Contains(aPairBox, pairB.BaseAsset) || slices.Contains(aPairBox, pairB.QuoteAsset) {
				// get pair C
				for _, pairC := range pairsList {
					if pairC == pairA || pairC == pairB || pairA == pairB {
						continue
					}

					pairBox := []string{
						pairA.BaseAsset,
						pairA.QuoteAsset,
						pairB.BaseAsset,
						pairB.QuoteAsset,
						pairC.BaseAsset,
						pairC.QuoteAsset,
					}

					var countsCBase = 0
					var countsCQuote = 0

					for _, value := range pairBox {
						if value == pairC.BaseAsset {
							countsCBase++
						}
						if value == pairC.QuoteAsset {
							countsCQuote++
						}
					}

					// found a triangular match
					if pairC.BaseAsset != pairC.QuoteAsset && countsCBase == 2 && countsCQuote == 2 {
						var pairSymbols = []string{pairA.Symbol, pairB.Symbol, pairC.Symbol}
						sort.Slice(pairSymbols, func(i, j int) bool {
							return pairSymbols[i] < pairSymbols[j]
						})
						var key = strings.Join(pairSymbols, "_")

						if _, ok := removeDuplicatesMap[key]; !ok {
							triangularPairsList = append(triangularPairsList, [3]*sourceprovider.Symbol{
								pairA,
								pairB,
								pairC,
							})
							removeDuplicatesMap[key] = true
						}
					}
				}
			}
		}
	}

	return triangularPairsList
}
