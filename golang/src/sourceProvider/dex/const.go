package dex

import (
	"arbitrage-bot/sourceprovider"
	"time"
)

// SymbolPrice ... Represents the price of a symbol in a DEX
type SymbolPrice struct {
	Symbol      *sourceprovider.Symbol `json:"symbol"`
	Token0Price float64                `json:"token0Price"`
	Token1Price float64                `json:"token1Price"`
	EventTime   time.Time              `json:"eventTime"`
}

// UniswapGraphQLURL ... Uniswap GraphQL endpoint
const UniswapGraphQLURL string = "https://gateway.thegraph.com/api/{api-key}/subgraphs/id/5zvR82QoaXYFyDEKLZ9t6v9adgnptxYpKpSbxtgVENFV"

// SubgraphPoolItem ... Uniswap subgraph pool item
type SubgraphPoolItem struct {
	ID                  string `json:"id"`
	FeeTier             string `json:"feeTier"`
	TotalValueLockedETH string `json:"totalValueLockedETH"`
	Token0Price         string `json:"token0Price"`
	Token1Price         string `json:"token1Price"`
	Token0              struct {
		Decimals string `json:"decimals"`
		ID       string `json:"id"`
		Symbol   string `json:"symbol"`
	} `json:"token0"`
	Token1 struct {
		Decimals string `json:"decimals"`
		ID       string `json:"id"`
		Symbol   string `json:"symbol"`
	} `json:"token1"`
}

// UniswapTokenListPath ... Uniswap token list path
const UniswapTokenListPath string = "data/uniswapTokens.json"

// UniswapArbitragePairPath ... Uniswap arbitrage pair path
const UniswapArbitragePairPath string = "data/uniswapArbitragePairs.json"

// ISourceProvider ... Interface for the DEX source provider
type ISourceProvider interface {
	sourceprovider.ISourceProvider
	GetSymbolPrice(symbol string) *SymbolPrice
}
