package dex

import (
	"arbitrage-bot/helpers"
	ioHelper "arbitrage-bot/helpers/io"
	"arbitrage-bot/sourceprovider"
	"encoding/json"
)

const surfaceRateQuery string = `
{
	pools(
		orderBy:totalValueLockedETH,
		orderDirection: desc,
		first: 500,
		skip: 4
	) {
		token0 {id symbol decimals}
		token1 {id symbol decimals}
		id
		totalValueLockedETH
		token1Price
		token0Price
		feeTier
	}
}`

// UniswapSourceProvider ... Uniswap source provider
type UniswapSourceProvider struct{}

// NewUniswapSourceProvider ... creates a new Uniswap source provider
func NewUniswapSourceProvider() *UniswapSourceProvider {
	return &UniswapSourceProvider{}
}

// GetArbitragePairCachePath ... returns the path to the token list cache
func (u *UniswapSourceProvider) GetArbitragePairCachePath() string {
	return UniswapArbitragePairPath
}

// GetTokenListCachePath ... returns the path to the token list cache
func (u *UniswapSourceProvider) GetTokenListCachePath() string {
	return UniswapTokenListPath
}

// GetSymbolPrice ... returns the price for a given symbol
func (u *UniswapSourceProvider) GetSymbolPrice(symbol string) *sourceprovider.SymbolPrice {
	return nil
}

// GetSymbolOrderbookDepth ... returns the order book for a given symbol
func (u *UniswapSourceProvider) GetSymbolOrderbookDepth(symbol string) *sourceprovider.SymbolOrderbookDepth {
	return nil
}

// GetSymbols ... returns the symbols
func (u *UniswapSourceProvider) GetSymbols(force bool) ([]*sourceprovider.Symbol, error) {
	requestBody, err := json.Marshal(map[string]string{"query": surfaceRateQuery})
	helpers.Panic(err)
	resData, err := ioHelper.Post(UniswapGraphQLURL, requestBody)
	helpers.Panic(err)

	resDataItems := (*resData)["data"].(map[string]interface{})["pools"].([]interface{})
	subgraphPoolItems := make([]SubgraphPoolItem, len(resDataItems))

	for i, item := range resDataItems {
		itemMap := item.(map[string]interface{})
		itemJSON, err := json.Marshal(itemMap)
		helpers.Panic(err)
		err = json.Unmarshal(itemJSON, &subgraphPoolItems[i])
		helpers.Panic(err)
	}

	symbols := make([]*sourceprovider.Symbol, len(subgraphPoolItems))

	for i, item := range subgraphPoolItems {
		symbols[i] = &sourceprovider.Symbol{
			Symbol:     item.Token0.Symbol + item.Token1.Symbol,
			BaseAsset:  item.Token0.Symbol,
			QuoteAsset: item.Token1.Symbol,
		}
	}

	return symbols, nil
}

// SubscribeSymbols ... subscribes to the symbols
func (u *UniswapSourceProvider) SubscribeSymbols(symbols []*sourceprovider.Symbol) {

}
