package dex

import (
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/web3"
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

// ISourceProvider ... Interface for the DEX source provider
type ISourceProvider interface {
	Web3Service() web3.DEXWeb3Service
	sourceprovider.ISourceProvider
	SubscribeSymbols(symbols []*sourceprovider.Symbol, pingChannel chan bool)
	GetSymbol(symbol string) sourceprovider.Symbol
	GetSymbolPrice(symbol string) *SymbolPrice
}
