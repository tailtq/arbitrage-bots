package arbitrage

type TriangularBidAskPrice struct {
	pairAAsk float64
	pairABid float64
	pairBAsk float64
	pairBBid float64
	pairCAsk float64
	pairCBid float64
}

type TriangularDexPrice struct {
	pairAToken0 float64
	pairAToken1 float64
	pairBToken0 float64
	pairBToken1 float64
	pairCToken0 float64
	pairCToken1 float64
}

const MinSurfaceRate float64 = 0.0 // the rate that indicates the arbitrage is profitable or not (and to prevent tiny wins)
