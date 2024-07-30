package arbitrage

import (
	"arbitrage-bot/sourceProvider"
	"slices"
	"sort"
	"strings"
)

type TriangularPairFinder struct{}

func (t *TriangularPairFinder) Handle(symbols []*sourceProvider.Symbol, amount int) [][3]*sourceProvider.Symbol {
	// find a list of 3 arbitrage pairs (f.e. SEIBNB BNBBTC SEIBTC)
	var triangularPairsList [][3]*sourceProvider.Symbol
	var removeDuplicatesMap map[string]bool = make(map[string]bool)
	var pairsList []*sourceProvider.Symbol = symbols[:amount]

	// get pair A
	for _, pairA := range pairsList {
		var aPairBox []string = []string{pairA.BaseAsset, pairA.QuoteAsset}

		// get pair B
		for _, pairB := range pairsList {
			if pairA == pairB {
				continue
			}

			// if three pairs form a cycle, continue
			if slices.Contains(aPairBox, pairB.BaseAsset) || slices.Contains(aPairBox, pairB.QuoteAsset) {
				var abPairBox []string = append([]string{pairB.BaseAsset, pairB.QuoteAsset}, aPairBox...)

				// get pair C
				for _, pairC := range pairsList {
					if pairC == pairA || pairC == pairB {
						continue
					}

					// found a triangular match
					if slices.Contains(abPairBox, pairC.BaseAsset) && slices.Contains(abPairBox, pairC.QuoteAsset) {
						var pairSymbols []string = []string{pairA.Symbol, pairB.Symbol, pairC.Symbol}
						sort.Slice(pairSymbols, func(i, j int) bool {
							return pairSymbols[i] < pairSymbols[j]
						})
						var key string = strings.Join(pairSymbols, "_")

						if _, ok := removeDuplicatesMap[key]; !ok {
							triangularPairsList = append(triangularPairsList, [3]*sourceProvider.Symbol{
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
