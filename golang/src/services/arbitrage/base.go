package arbitrage

import (
	"arbitrage-bot/models"
	"arbitrage-bot/services/sourceprovider"
)

type IArbitrageCalculator interface {
	NewArbitrageCalculator(sourceProvider sourceprovider.ISourceProvider) IArbitrageCalculator
	CalcTriangularArbSurfaceRate(
		triangularPair [3]*sourceprovider.Symbol, startingAmount float64,
	) (models.TriangularArbSurfaceResult, error)
	GetDepth(surfaceRate models.TriangularArbSurfaceResult) models.TriangularArbDepthResult
}
