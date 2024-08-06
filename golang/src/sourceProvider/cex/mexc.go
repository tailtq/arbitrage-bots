package CEX

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/sourceProvider"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MEXCSourceProvider struct {
	// data stream
	streamsTicker         []*ioHelper.WebSocketClient
	streamsOrderbookDepth []*ioHelper.WebSocketClient
	symbols               map[string]*sourceProvider.Symbol
	symbolPriceData       sync.Map
	symbolOrderbookData   sync.Map
}

// NewMEXCSourceProvider ... creates a new MEXC source provider
func NewMEXCSourceProvider() *MEXCSourceProvider {
	return &MEXCSourceProvider{
		symbols: make(map[string]*sourceProvider.Symbol),
	}
}

// GetArbitragePairCachePath implements sourceProvider.SourceProviderInterface.
func (b *MEXCSourceProvider) GetArbitragePairCachePath() string {
	return MEXCArbitragePairPath
}

// GetTokenListCachePath implements sourceProvider.SourceProviderInterface.
func (b *MEXCSourceProvider) GetTokenListCachePath() string {
	return MEXCTokenListPath
}

// GetSymbolPrice returns the price for a given symbol
func (b *MEXCSourceProvider) GetSymbolPrice(symbol string) *sourceProvider.SymbolPrice {
	if price, ok := b.symbolPriceData.Load(symbol); ok {
		return price.(*sourceProvider.SymbolPrice)
	}

	return nil
}

// GetSymbolOrderbookDepth ... returns the order book for a given symbol
func (b *MEXCSourceProvider) GetSymbolOrderbookDepth(symbol string) *sourceProvider.SymbolOrderbookDepth {
	if orderbook, ok := b.symbolOrderbookData.Load(symbol); ok {
		return orderbook.(*sourceProvider.SymbolOrderbookDepth)
	}

	return nil
}

// GetSymbols ... returns a list of symbols
func (b *MEXCSourceProvider) GetSymbols(force bool) ([]*sourceProvider.Symbol, error) {
	// get a list of all symbols on Binance & save to file as cache
	if !force && fileHelper.PathExists(MEXCTokenListPath) {
		var symbols []*sourceProvider.Symbol
		err := jsonHelper.ReadJSONFile(MEXCTokenListPath, &symbols)

		return symbols, err
	}
	var data *map[string]interface{}
	data, err := ioHelper.Get(MEXCApiUrl+"/exchangeInfo", data)
	helpers.Panic(err)

	dataMap := make([]*sourceProvider.Symbol, 0)
	// Type assertion (a way to retrieve the dynamic type of an interface)
	symbols, ok := (*data)["symbols"].([]interface{})

	if ok {
		for _, symbol := range symbols {
			s, ok := symbol.(map[string]interface{})
			var quoteAsset string = s["quoteAsset"].(string)

			// skip other USD assets but USDT and USDC because they don't seem to be reliable
			if ok && strings.Contains(quoteAsset, "USD") && quoteAsset != "USDT" && quoteAsset != "USDC" {
				continue
			}

			if ok && s["isSpotTradingAllowed"].(bool) {
				dataMap = append(dataMap, &sourceProvider.Symbol{
					Symbol:     s["symbol"].(string),
					BaseAsset:  s["baseAsset"].(string),
					QuoteAsset: s["quoteAsset"].(string),
				})
				fmt.Println("dataMap", *dataMap[len(dataMap)-1])
			}
		}
		// save to file
		jsonHelper.WriteJSONFile(MEXCTokenListPath, symbols)
	}

	return dataMap, err
}

// SubscribeSymbols ... subscribes to a list of symbols
func (b *MEXCSourceProvider) SubscribeSymbols(symbols []*sourceProvider.Symbol) {
	// subscribe a new data stream for a new symbol
	// check if symbol already exists
	for _, symbol := range symbols {
		if _, ok := b.symbols[symbol.Symbol]; ok {
			continue
		}

		b.symbols[symbol.Symbol] = symbol
	}
	b.stopTickerDataStream()
	b.stopOrderbookDepthStream()
	b.startTickerDataStream()
	b.startOrderbookDepthStream()
}

func (b *MEXCSourceProvider) startTickerDataStream() {
	// subscribe to multiple data streams using one connection (ticker topic)
	// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-ticker-streams
	var symbols []string

	for symbol := range b.symbols {
		symbols = append(symbols, "spot@public.bookTicker.v3.api@"+symbol)
	}

	for _, paramBatch := range helpers.Batch(symbols, 24) {
		subscriptionEvent := MEXCEventSubscriptionUnsubscription{
			Method: "SUBSCRIPTION",
			Params: paramBatch,
		}
		streamTicker := ioHelper.NewWebSocketClient(MEXCWsUrl)
		streamTicker.Start(b.handleTickerDataStream)
		streamTicker.WriteJSON(subscriptionEvent)
		b.streamsTicker = append(b.streamsTicker, streamTicker)
	}
}

func (b *MEXCSourceProvider) handleTickerDataStream(data *[]byte) {
	var ticker MEXCSymbolTicker
	jsonHelper.Unmarshal(*data, &ticker)
	bestAsk, _ := strconv.ParseFloat(ticker.Data.BestAskPrice, 64)
	bestBid, _ := strconv.ParseFloat(ticker.Data.BestBidPrice, 64)
	bestAskQuantity, _ := strconv.ParseFloat(ticker.Data.BestAskQuantity, 64)
	bestBidQuantity, _ := strconv.ParseFloat(ticker.Data.BestBidQuantity, 64)

	b.symbolPriceData.Store(ticker.Symbol, &sourceProvider.SymbolPrice{
		Symbol:          b.symbols[ticker.Symbol],
		BestBid:         bestBid,
		BestBidQuantity: bestBidQuantity,
		BestAsk:         bestAsk,
		BestAskQuantity: bestAskQuantity,
		EventTime:       time.Unix(0, ticker.Time*1000000),
	})
}

func (b *MEXCSourceProvider) stopTickerDataStream() {
	for _, streamTicker := range b.streamsTicker {
		streamTicker.Stop()
	}
}

func (b *MEXCSourceProvider) UnsubscribeSymbol(symbol *sourceProvider.Symbol) {
	delete(b.symbols, symbol.Symbol)
	b.stopTickerDataStream()
	b.stopOrderbookDepthStream()
	b.startTickerDataStream()
	b.startOrderbookDepthStream()
}

func (b *MEXCSourceProvider) startOrderbookDepthStream() {
	// Stop because the arbitrage rate is negative
}

func (b *MEXCSourceProvider) handleOrderbookDepthStream(data *[]byte) {

}

func (b *MEXCSourceProvider) stopOrderbookDepthStream() {

}
