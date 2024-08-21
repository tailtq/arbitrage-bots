package arbitrage

import (
	sp "arbitrage-bot/sourceprovider"
)

type IArbitrageCalculator interface {
	NewArbitrageCalculator(sourceProvider sp.ISourceProvider) IArbitrageCalculator
	CalcTriangularArbSurfaceRate(triangularPair [3]*sp.Symbol, startingAmount float64) (TriangularArbSurfaceResult, error)
	GetDepthFromOrderBook(surfaceRate TriangularArbSurfaceResult) TriangularArbDepthResult
}
