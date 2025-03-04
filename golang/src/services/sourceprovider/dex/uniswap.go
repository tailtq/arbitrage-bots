package dex

import (
	"arbitrage-bot/helpers"
	ioHelper "arbitrage-bot/helpers/io"
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/web3"
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
)

const surfaceRateQuery string = `
{
	pools(
		orderBy:totalValueLockedETH,
		orderDirection: desc,
		first: 1000,
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

var suitablePairs = []string{
	"DAI", "FRAX", "FEI", "LINK", "DAI", "PEPE", "DYAD", "MKR", "UNI", "WHITE", "MNT", "TURBO", "SHIB", "DOG", "APE", "ENS", "PANDORA", "SOL", "LDO", "MATIC", "AAVE", "ONDO", "PEOPLE", "SHFL", "FTM", "RNDR", "KOIN", "RCH", "FET", "LBTC", "PORK", "PRIME", "HEX",
}

// UniswapSourceProviderService ... Uniswap source provider
type UniswapSourceProviderService struct {
	web3Service     web3.DEXWeb3Service
	symbolPriceData sync.Map
	symbols         map[string]*sourceprovider.Symbol
}

// NewUniswapSourceProviderService ... creates a new Uniswap source provider
func NewUniswapSourceProviderService() *UniswapSourceProviderService {
	return &UniswapSourceProviderService{
		symbols:     make(map[string]*sourceprovider.Symbol),
		web3Service: web3.NewUniswapWeb3Service(),
	}
}

func (u *UniswapSourceProviderService) Web3Service() web3.DEXWeb3Service {
	return u.web3Service
}

// GetArbitragePairCachePath ... returns the path to the token list cache
func (u *UniswapSourceProviderService) GetArbitragePairCachePath() string {
	var network = os.Getenv("NETWORK_NAME")
	return "data/" + network + "/uniswapArbitragePairs.json"
}

// GetSymbolPrice ... returns the aggregated price for a given symbol
func (u *UniswapSourceProviderService) GetSymbolPrice(symbol string) *SymbolPrice {
	if price, ok := u.symbolPriceData.Load(symbol); ok {
		return price.(*SymbolPrice)
	}
	return nil
}

// getSubgraphPoolData ... gets depth price data from the subgraph pool
func (u *UniswapSourceProviderService) getSubgraphPoolData() ([]SubgraphPoolItem, error) {
	requestBody, err := json.Marshal(map[string]string{"query": surfaceRateQuery})

	if err != nil {
		return nil, err
	}

	resData := make(map[string]interface{})
	err = ioHelper.Post(UniswapGraphQLURL(), requestBody, &resData)

	if err != nil {
		return nil, err
	}

	resDataItems := resData["data"].(map[string]interface{})["pools"].([]interface{})
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

// GetSymbol ... returns the symbol
func (u *UniswapSourceProviderService) GetSymbol(symbol string) sourceprovider.Symbol {
	return *u.symbols[symbol]
}

// GetSymbols ... returns the symbols
func (u *UniswapSourceProviderService) GetSymbols(force bool) ([]*sourceprovider.Symbol, error) {
	subgraphPoolItems, err := u.getSubgraphPoolData()
	helpers.Panic(err)
	var symbols []*sourceprovider.Symbol
	uniqueSymbols := make(map[string]bool)

	for _, item := range subgraphPoolItems {
		feeTier, _ := strconv.Atoi(item.FeeTier)
		baseAssetDecimals, _ := strconv.Atoi(item.Token0.Decimals)
		quoteAssetDecimals, _ := strconv.Atoi(item.Token1.Decimals)
		pair := item.Token0.Symbol + item.Token1.Symbol

		if uniqueSymbols[pair] || (!slices.Contains(suitablePairs, item.Token0.Symbol) &&
			!slices.Contains(suitablePairs, item.Token1.Symbol)) {
			continue
		}

		symbols = append(symbols, &sourceprovider.Symbol{
			Address:            item.ID,
			Symbol:             pair,
			FeeTier:            feeTier,
			BaseAsset:          item.Token0.Symbol,
			BaseAssetAddress:   item.Token0.ID,
			BaseAssetDecimals:  baseAssetDecimals,
			QuoteAsset:         item.Token1.Symbol,
			QuoteAssetAddress:  item.Token1.ID,
			QuoteAssetDecimals: quoteAssetDecimals,
		})
		uniqueSymbols[pair] = true
	}

	return symbols, nil
}

// SubscribeSymbols ... subscribes to the symbols
func (u *UniswapSourceProviderService) SubscribeSymbols(
	symbols []*sourceprovider.Symbol, pingChannel chan bool, verbose bool,
) {
	var tokenPairs []string

	for _, symbol := range symbols {
		u.symbols[symbol.Symbol] = symbol
		tokenPairs = append(tokenPairs, symbol.Symbol)
	}
	for {
		// Fetch the data directly from the network
		aggregatedPrices := u.web3Service.AggregatePrices(symbols, verbose)
		aggregatedPrices.Range(func(key any, value any) bool {
			u.symbolPriceData.Store(key, &SymbolPrice{
				Symbol:      u.symbols[key.(string)],
				Token0Price: 1.0 / value.(float64),
				Token1Price: value.(float64),
				EventTime:   time.Now(),
			})
			return true
		})
		pingChannel <- true

		// Fetch the data every 60 seconds
		time.Sleep(10 * time.Second)
	}
}
