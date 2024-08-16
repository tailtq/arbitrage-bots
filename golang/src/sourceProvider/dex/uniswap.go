package dex

import (
	"arbitrage-bot/helpers"
	ioHelper "arbitrage-bot/helpers/io"
	"arbitrage-bot/sourceprovider"
	"encoding/json"
	"strconv"
	"sync"
	"time"
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
type UniswapSourceProvider struct {
	symbolPriceData sync.Map
	symbols         map[string]*sourceprovider.Symbol
}

// NewUniswapSourceProvider ... creates a new Uniswap source provider
func NewUniswapSourceProvider() *UniswapSourceProvider {
	return &UniswapSourceProvider{
		symbols: make(map[string]*sourceprovider.Symbol),
	}
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
func (u *UniswapSourceProvider) GetSymbolPrice(symbol string) *SymbolPrice {
	if price, ok := u.symbolPriceData.Load(symbol); ok {
		return price.(*SymbolPrice)
	}

	return nil
}

// GetSymbolOrderbookDepth ... returns the order book for a given symbol
func (u *UniswapSourceProvider) GetSymbolOrderbookDepth(symbol string) *sourceprovider.SymbolOrderbookDepth {
	return nil
}

func (u *UniswapSourceProvider) getSubgraphPoolData() ([]SubgraphPoolItem, error) {
	requestBody, err := json.Marshal(map[string]string{"query": surfaceRateQuery})

	if err != nil {
		return nil, err
	}

	resData, err := ioHelper.Post(UniswapGraphQLURL, requestBody)

	if err != nil {
		return nil, err
	}

	resDataItems := (*resData)["data"].(map[string]interface{})["pools"].([]interface{})
	subgraphPoolItems := make([]SubgraphPoolItem, len(resDataItems))

	for i, item := range resDataItems {
		itemMap := item.(map[string]interface{})
		itemJSON, err := json.Marshal(itemMap)

		if err == nil {
			err = json.Unmarshal(itemJSON, &subgraphPoolItems[i])
		}
		if err != nil {
			return nil, err
		}
	}

	return subgraphPoolItems, nil
}

// GetSymbols ... returns the symbols
func (u *UniswapSourceProvider) GetSymbols(force bool) ([]*sourceprovider.Symbol, error) {
	subgraphPoolItems, err := u.getSubgraphPoolData()
	helpers.Panic(err)
	var symbols []*sourceprovider.Symbol
	uniqueSymbols := make(map[string]bool)

	for _, item := range subgraphPoolItems {
		baseAssetDecimals, _ := strconv.Atoi(item.Token0.Decimals)
		quoteAssetDecimals, _ := strconv.Atoi(item.Token1.Decimals)
		pair := item.Token0.Symbol + item.Token1.Symbol

		if uniqueSymbols[pair] {
			continue
		}

		symbols = append(symbols, &sourceprovider.Symbol{
			Symbol:             pair,
			BaseAsset:          item.Token0.Symbol,
			BaseAssetID:        item.Token0.ID,
			BaseAssetDecimals:  uint8(baseAssetDecimals),
			QuoteAsset:         item.Token1.Symbol,
			QuoteAssetID:       item.Token1.ID,
			QuoteAssetDecimals: uint8(quoteAssetDecimals),
		})
		uniqueSymbols[pair] = true
	}

	return symbols, nil
}

// SubscribeSymbols ... subscribes to the symbols
func (u *UniswapSourceProvider) SubscribeSymbols(symbols []*sourceprovider.Symbol) {
	for _, symbol := range symbols {
		u.symbols[symbol.Symbol] = symbol
	}

	for {
		subgraphPoolItems, err := u.getSubgraphPoolData()
		helpers.Panic(err)

		for _, item := range subgraphPoolItems {
			token0Price, _ := strconv.ParseFloat(item.Token0Price, 64)
			token1Price, _ := strconv.ParseFloat(item.Token1Price, 64)
			symbol := u.symbols[item.Token0.Symbol+item.Token1.Symbol]

			if symbol == nil {
				continue
			}

			u.symbolPriceData.Store(symbol.Symbol, &SymbolPrice{
				Symbol:      symbol,
				Token0Price: token0Price,
				Token1Price: token1Price,
				EventTime:   time.Now(),
			})
		}

		// Fetch the data every 60 seconds
		time.Sleep(60 * time.Second)
	}
}
