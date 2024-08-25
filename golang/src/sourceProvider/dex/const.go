package dex

import (
	"arbitrage-bot/models"
	"arbitrage-bot/sourceprovider"
	"os"
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
func UniswapGraphQLURL() string {
	return "https://gateway.thegraph.com/api/" + os.Getenv("SUBGRAPH_API_KEY") + "/subgraphs/id/" + os.Getenv("SUBGRAPH_UNISWAP_ID")
}

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
	GetDepth(surfaceRate models.TriangularArbSurfaceResult) ([2]models.TriangularArbDepthResult, error)
}
