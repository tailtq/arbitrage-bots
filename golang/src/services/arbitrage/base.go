package arbitrage

import (
	"arbitrage-bot/models"
	sp "arbitrage-bot/sourceprovider"
)

type IArbitrageCalculator interface {
	NewArbitrageCalculator(sourceProvider sp.ISourceProvider) IArbitrageCalculator
	CalcTriangularArbSurfaceRate(
		triangularPair [3]*sp.Symbol, startingAmount float64,
	) (models.TriangularArbSurfaceResult, error)
	GetDepth(surfaceRate models.TriangularArbSurfaceResult) models.TriangularArbDepthResult
}
