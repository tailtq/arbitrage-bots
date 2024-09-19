package models

import "arbitrage-bot/services/sourceprovider"

type TriangularArbSurfaceResult struct {
	Swap1             string                `json:"swap1"`
	Swap2             string                `json:"swap2"`
	Swap3             string                `json:"swap3"`
	Contract1         string                `json:"contract1"`
	Contract2         string                `json:"contract2"`
	Contract3         string                `json:"contract3"`
	Symbol1           sourceprovider.Symbol `json:"symbol1"`
	Symbol2           sourceprovider.Symbol `json:"symbol2"`
	Symbol3           sourceprovider.Symbol `json:"symbol3"`
	Contract1Address  string                `json:"contract1Address"`
	Contract2Address  string                `json:"contract2Address"`
	Contract3Address  string                `json:"contract3Address"`
	DirectionTrade1   string                `json:"directionTrade1"`
	DirectionTrade2   string                `json:"directionTrade2"`
	DirectionTrade3   string                `json:"directionTrade3"`
	StartingAmount    float64               `json:"startingAmount"`
	AcquiredCoinT1    float64               `json:"acquiredCoinT1"`
	AcquiredCoinT2    float64               `json:"acquiredCoinT2"`
	AcquiredCoinT3    float64               `json:"acquiredCoinT3"`
	Swap1Rate         float64               `json:"swap1Rate"`
	Swap2Rate         float64               `json:"swap2Rate"`
	Swap3Rate         float64               `json:"swap3Rate"`
	ProfitLoss        float64               `json:"profitLoss"`
	ProfitLossPerc    float64               `json:"profitLossPerc"`
	Direction         string                `json:"direction"`
	TradeDescription1 string                `json:"tradeDescription1"`
	TradeDescription2 string                `json:"tradeDescription2"`
	TradeDescription3 string                `json:"tradeDescription3"`
}

type TriangularArbDepthResult struct {
	ProfitLoss     float64
	ProfitLossPerc float64
}

type TriangularArbFullResult struct {
	SurfaceResult      TriangularArbSurfaceResult
	DepthResultForward TriangularArbDepthResult
}
