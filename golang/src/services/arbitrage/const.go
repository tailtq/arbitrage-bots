package arbitrage

type TriangularBidAskPrice struct {
	pairAAsk float64
	pairABid float64
	pairBAsk float64
	pairBBid float64
	pairCAsk float64
	pairCBid float64
}

type TriangularSurfaceRate struct {
	Swap1             string
	Swap2             string
	Swap3             string
	Contract1         string
	Contract2         string
	Contract3         string
	DirectionTrade1   string
	DirectionTrade2   string
	DirectionTrade3   string
	StartingAmount    float64
	AcquiredCoinT1    float64
	AcquiredCoinT2    float64
	AcquiredCoinT3    float64
	Swap1Rate         float64
	Swap2Rate         float64
	Swap3Rate         float64
	ProfitLoss        float64
	ProfitLossPerc    float32
	Direction         string
	TradeDescription1 string
	TradeDescription2 string
	TradeDescription3 string
}

const MinSurfaceRate float64 = 0.01 // the rate that indicates the arbitrage is profitable or not (and to prevent tiny wins)
