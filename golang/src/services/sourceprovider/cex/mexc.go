package cex

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/sourceprovider"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MEXCSourceProviderService ...
type MEXCSourceProviderService struct {
	// data stream
	streamsTicker         []*ioHelper.WebSocketClient
	streamsOrderbookDepth []*ioHelper.WebSocketClient
	symbols               map[string]*sourceprovider.Symbol
	symbolPriceData       sync.Map
	symbolOrderbookData   sync.Map
}

// NewMEXCSourceProviderService ... creates a new MEXC source provider
func NewMEXCSourceProviderService() *MEXCSourceProviderService {
	return &MEXCSourceProviderService{
		symbols: make(map[string]*sourceprovider.Symbol),
	}
}

// GetArbitragePairCachePath implements sourceprovider.ICexSourceProvider.
func (b *MEXCSourceProviderService) GetArbitragePairCachePath() string {
	return MEXCArbitragePairPath
}

// GetSymbolPrice returns the price for a given symbol
func (b *MEXCSourceProviderService) GetSymbolPrice(symbol string) *SymbolPrice {
	if price, ok := b.symbolPriceData.Load(symbol); ok {
		return price.(*SymbolPrice)
	}

	return nil
}

// GetSymbolOrderbookDepth ... returns the order book for a given symbol
func (b *MEXCSourceProviderService) GetSymbolOrderbookDepth(symbol string) *sourceprovider.SymbolOrderbookDepth {
	if orderbook, ok := b.symbolOrderbookData.Load(symbol); ok {
		return orderbook.(*sourceprovider.SymbolOrderbookDepth)
	}

	return nil
}

// GetSymbols ... returns a list of symbols
func (b *MEXCSourceProviderService) GetSymbols(force bool) ([]*sourceprovider.Symbol, error) {
	// get a list of all symbols on Binance & save to file as cache
	if !force && fileHelper.PathExists(MEXCTokenListPath) {
		var symbols []*sourceprovider.Symbol
		err := jsonHelper.ReadJSONFile(MEXCTokenListPath, &symbols)

		return symbols, err
	}
	var data = make(map[string]interface{})
	err := ioHelper.Get(MEXCAPIURL+"/exchangeInfo", &data)
	helpers.Panic(err)

	dataMap := make([]*sourceprovider.Symbol, 0)
	// Type assertion (a way to retrieve the dynamic type of an interface)
	symbols, ok := data["symbols"].([]interface{})

	if ok {
		for _, symbol := range symbols {
			s, ok := symbol.(map[string]interface{})
			var quoteAsset string = s["quoteAsset"].(string)

			// skip other USD assets but USDT and USDC because they don't seem to be reliable
			if ok && strings.Contains(quoteAsset, "USD") && quoteAsset != "USDT" && quoteAsset != "USDC" {
				continue
			}

			if ok && s["isSpotTradingAllowed"].(bool) {
				dataMap = append(dataMap, &sourceprovider.Symbol{
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
func (b *MEXCSourceProviderService) SubscribeSymbols(symbols []*sourceprovider.Symbol) {
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

func (b *MEXCSourceProviderService) startTickerDataStream() {
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
		streamTicker := ioHelper.NewWebSocketClient(MEXCWsURL)
		streamTicker.Start(b.handleTickerDataStream)
		streamTicker.WriteJSON(subscriptionEvent)
		b.streamsTicker = append(b.streamsTicker, streamTicker)
	}
}

func (b *MEXCSourceProviderService) handleTickerDataStream(data *[]byte) {
	var ticker MEXCSymbolTicker
	jsonHelper.Unmarshal(*data, &ticker)
	bestAsk, _ := strconv.ParseFloat(ticker.Data.BestAskPrice, 64)
	bestBid, _ := strconv.ParseFloat(ticker.Data.BestBidPrice, 64)

	b.symbolPriceData.Store(ticker.Symbol, &SymbolPrice{
		Symbol:    b.symbols[ticker.Symbol],
		BestBid:   bestBid,
		BestAsk:   bestAsk,
		EventTime: time.Unix(0, ticker.Time*1000000),
	})
}

func (b *MEXCSourceProviderService) stopTickerDataStream() {
	for _, streamTicker := range b.streamsTicker {
		streamTicker.Stop()
	}
}

// UnsubscribeSymbol ... unsubscribes from a symbol
func (b *MEXCSourceProviderService) UnsubscribeSymbol(symbol *sourceprovider.Symbol) {
	delete(b.symbols, symbol.Symbol)
	b.stopTickerDataStream()
	b.stopOrderbookDepthStream()
	b.startTickerDataStream()
	b.startOrderbookDepthStream()
}

func (b *MEXCSourceProviderService) startOrderbookDepthStream() {
	// Stop because the arbitrage rate is negative
}

func (b *MEXCSourceProviderService) handleOrderbookDepthStream(data *[]byte) {
	// Stop because the arbitrage rate is negative
}

func (b *MEXCSourceProviderService) stopOrderbookDepthStream() {
	// Stop because the arbitrage rate is negative
}
