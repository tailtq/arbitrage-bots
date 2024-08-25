package dex

import (
	"arbitrage-bot/helpers"
	ioHelper "arbitrage-bot/helpers/io"
	"arbitrage-bot/models"
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
		first: 700,
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

	resData := make(map[string]interface{})
	err = ioHelper.Post(UniswapGraphQLURL(), requestBody, &resData)

	if err != nil {
		return nil, err
	}

	//fmt.Println("resData", resData)
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

// loadTokens ... loads the tokens for calculating depths before running
func (u *UniswapSourceProvider) loadTokens(addresses []string) {
	requestBody, err := json.Marshal(map[string][]string{
		"pairAddresses": addresses,
	})
	helpers.Panic(err)

	var responseData []interface{}
	var uniswapLoadTokensAPI = helpers.GetEnv("UNISWAP_NODEJS_SERVER") + "/uniswap/tokens/load"
	err = ioHelper.Post(uniswapLoadTokensAPI, requestBody, &responseData)
	helpers.Panic(err)
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
			ID:                 item.ID,
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
	var symbolAddresses []string

	for _, symbol := range symbols {
		u.symbols[symbol.Symbol] = symbol
		symbolAddresses = append(symbolAddresses, symbol.ID)
	}

	u.loadTokens(symbolAddresses)

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
		time.Sleep(10 * time.Second)
	}
}

// GetDepth ... returns the depth for a given surface rate
func (u *UniswapSourceProvider) GetDepth(surfaceRate models.TriangularArbSurfaceResult) ([2]models.TriangularArbDepthResult, error) {
	results, err := u.BatchGetDepth([]models.TriangularArbSurfaceResult{surfaceRate})

	if err != nil {
		return [2]models.TriangularArbDepthResult{}, err
	}

	return results[0], nil
}

func (u *UniswapSourceProvider) BatchGetDepth(surfaceRates []models.TriangularArbSurfaceResult) ([][2]models.TriangularArbDepthResult, error) {
	var results [][2]models.TriangularArbDepthResult
	var uniswapDepthAPI = helpers.GetEnv("UNISWAP_NODEJS_SERVER") + "/uniswap/arbitrage/batch-depth"
	requestBody, err := json.Marshal(map[string]any{"surfaceResults": surfaceRates})
	if err != nil {
		return results, err
	}

	var responseBatchData = make(map[string]interface{})
	err = ioHelper.Post(uniswapDepthAPI, requestBody, &responseBatchData)
	if err != nil {
		return results, err
	}

	for _, surfaceRate := range surfaceRates {
		var key = surfaceRate.Swap1 + "_" + surfaceRate.Swap2 + "_" + surfaceRate.Swap3

		if _, ok := responseBatchData[key]; ok {
			var responseData = responseBatchData[key].(map[string]interface{})
			var resultItem [2]models.TriangularArbDepthResult

			if _, ok = responseData["forward"]; ok {
				var resultForward = responseData["forward"].(map[string]interface{})
				resultItem[0].ProfitLoss = resultForward["profitLoss"].(float64)
				resultItem[0].ProfitLossPerc = float32(resultForward["profitLossPerc"].(float64))
			}
			if _, ok = responseData["backward"]; ok {
				var resultBackward = responseData["backward"].(map[string]interface{})
				resultItem[1].ProfitLoss = resultBackward["profitLoss"].(float64)
				resultItem[1].ProfitLossPerc = float32(resultBackward["profitLossPerc"].(float64))
			}

			results = append(results, resultItem)
		}
	}

	return results, nil
}
