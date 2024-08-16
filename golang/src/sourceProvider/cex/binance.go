package cex

import (
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/sourceprovider"
)

// BinanceSourceProvider ... Binance source provider
type BinanceSourceProvider struct {
	// data stream
	streamTicker         *ioHelper.WebSocketClient
	streamOrderbookDepth *ioHelper.WebSocketClient
	symbols              map[string]*sourceprovider.Symbol
	// we'll get (fatal error: concurrent map read and map write) if using regular map
	symbolPriceData     sync.Map
	symbolOrderbookData sync.Map
}

// NewBinanceSourceProvider ... creates a new Binance source provider
func NewBinanceSourceProvider() *BinanceSourceProvider {
	return &BinanceSourceProvider{
		symbols: make(map[string]*sourceprovider.Symbol),
	}
}

// GetTokenListCachePath ... returns the path to the token list cache
func (b *BinanceSourceProvider) GetTokenListCachePath() string {
	return BinanceTokenListPath
}

// GetArbitragePairCachePath ... returns the path to the arbitrage pair cache
func (b *BinanceSourceProvider) GetArbitragePairCachePath() string {
	return BinanceArbitragePairPath
}

// GetSymbolPrice returns the price for a given symbol
func (b *BinanceSourceProvider) GetSymbolPrice(symbol string) *SymbolPrice {
	if price, ok := b.symbolPriceData.Load(symbol); ok {
		return price.(*SymbolPrice)
	}

	return nil
}

// GetSymbolOrderbookDepth ... returns the order book for a given symbol
func (b *BinanceSourceProvider) GetSymbolOrderbookDepth(symbol string) *sourceprovider.SymbolOrderbookDepth {
	if orderbook, ok := b.symbolOrderbookData.Load(symbol); ok {
		return orderbook.(*sourceprovider.SymbolOrderbookDepth)
	}

	return nil
}

// GetSymbols ... returns all the symbols
func (b *BinanceSourceProvider) GetSymbols(force bool) ([]*sourceprovider.Symbol, error) {
	// get a list of all symbols on Binance & save to file as cache
	if !force && fileHelper.PathExists(BinanceTokenListPath) {
		var symbols []*sourceprovider.Symbol
		err := jsonHelper.ReadJSONFile(BinanceTokenListPath, &symbols)

		return symbols, err
	}

	var data *map[string]interface{}
	data, err := ioHelper.Get(BinanceAPIURL+"/exchangeInfo", data)
	helpers.Panic(err)

	dataMap := make([]*sourceprovider.Symbol, 0)
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
				dataMap = append(dataMap, &sourceprovider.Symbol{
					Symbol:     s["symbol"].(string),
					BaseAsset:  s["baseAsset"].(string),
					QuoteAsset: s["quoteAsset"].(string),
				})
			}
		}
		// save to file
		jsonHelper.WriteJSONFile(BinanceTokenListPath, symbols)
	}

	return dataMap, err
}

// SubscribeSymbols ... subscribes to the symbols
func (b *BinanceSourceProvider) SubscribeSymbols(symbols []*sourceprovider.Symbol) {
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

func (b *BinanceSourceProvider) startTickerDataStream() {
	// subscribe to multiple data streams using one connection (ticker topic)
	// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-ticker-streams
	var symbolString string
	var charCount = 0

	for key := range b.symbols {
		symbolString += strings.ToLower(key) + "@ticker/"
	}

	if charCount = utf8.RuneCountInString(symbolString); charCount == 0 {
		return
	}

	symbolString = string([]rune(symbolString)[:charCount-1])
	var endpoint = BinanceWsURL + "/stream?streams=" + symbolString
	b.streamTicker = ioHelper.NewWebSocketClient(endpoint)
	b.streamTicker.Start(b.handleTickerDataStream)
}

func (b *BinanceSourceProvider) handleTickerDataStream(data *[]byte) {
	// handle the ticker data stream
	// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-ticker-streams
	var ticker BinanceSymbolTicker
	jsonHelper.Unmarshal(*data, &ticker)
	bestAsk, _ := strconv.ParseFloat(ticker.Data.BestAskPrice, 64)
	bestBid, _ := strconv.ParseFloat(ticker.Data.BestBidPrice, 64)

	b.symbolPriceData.Store(ticker.Data.Symbol, &SymbolPrice{
		Symbol:    b.symbols[ticker.Data.Symbol],
		BestBid:   bestBid,
		BestAsk:   bestAsk,
		EventTime: time.Unix(0, ticker.Data.EventTime*1000000),
	})

	// build a general interface so all exchanges can use the same data structure
	// fmt.Println(string(*data))
}

func (b *BinanceSourceProvider) stopTickerDataStream() {
	// stop the data stream if exists
	if b.streamTicker != nil {
		b.streamTicker.Stop()
	}
}

// UnsubscribeSymbol ... unsubscribes from the symbol
func (b *BinanceSourceProvider) UnsubscribeSymbol(symbol *sourceprovider.Symbol) {
	// unsubscribe a symbol from the data stream (remove symbol from the map -> restart data stream)
	delete(b.symbols, symbol.Symbol)
	b.stopTickerDataStream()
	b.stopOrderbookDepthStream()
	b.startTickerDataStream()
	b.startOrderbookDepthStream()
}

func (b *BinanceSourceProvider) startOrderbookDepthStream() {
	// subscribe to multiple data streams using one connection (order book/depth topic)
	var symbolString string
	var charCount int = 0

	for key := range b.symbols {
		symbolString += strings.ToLower(key) + "@depth20/"
	}

	if charCount = utf8.RuneCountInString(symbolString); charCount == 0 {
		return
	}

	symbolString = string([]rune(symbolString)[:charCount-1])
	var endpoint string = BinanceWsURL + "/stream?streams=" + symbolString
	b.streamOrderbookDepth = ioHelper.NewWebSocketClient(endpoint)
	b.streamOrderbookDepth.Start(b.handleOrderbookDepthStream)
}

func (b *BinanceSourceProvider) handleOrderbookDepthStream(data *[]byte) {
	// handle the ticker data stream
	var orderbookDepth BinanceOrderbookDepth
	jsonHelper.Unmarshal(*data, &orderbookDepth)
	var symbolOrderbookDepth sourceprovider.SymbolOrderbookDepth = sourceprovider.SymbolOrderbookDepth{
		Symbol:       b.symbols[orderbookDepth.GetSymbol()],
		LastUpdateID: orderbookDepth.Data.LastUpdateID,
		Asks:         make([]*sourceprovider.OrderbookEntry, len(orderbookDepth.Data.Asks)),
		Bids:         make([]*sourceprovider.OrderbookEntry, len(orderbookDepth.Data.Bids)),
	}

	for i, ask := range orderbookDepth.Data.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		quantity, _ := strconv.ParseFloat(ask[1], 64)
		symbolOrderbookDepth.Asks[i] = &sourceprovider.OrderbookEntry{
			Price:    price,
			Quantity: quantity,
		}
	}

	for i, bid := range orderbookDepth.Data.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		quantity, _ := strconv.ParseFloat(bid[1], 64)
		symbolOrderbookDepth.Bids[i] = &sourceprovider.OrderbookEntry{
			Price:    price,
			Quantity: quantity,
		}
	}

	b.symbolOrderbookData.Store(orderbookDepth.GetSymbol(), &symbolOrderbookDepth)

	// build a general interface so all exchanges can use the same data structure
	// fmt.Println(len(string(*data)), orderbookDepth.GetSymbol())
}

func (b *BinanceSourceProvider) stopOrderbookDepthStream() {
	// Get the depth from the order book
	if b.streamOrderbookDepth != nil {
		b.streamOrderbookDepth.Stop()
	}
}
