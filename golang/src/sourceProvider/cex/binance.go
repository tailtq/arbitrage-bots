package CEX

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/sourceProvider"
)

type BinanceSourceProvider struct {
	// data stream
	stream     *ioHelper.WebSocketClient
	symbols    map[string]*sourceProvider.Symbol
	symbolPriceData map[string]*sourceProvider.SymbolPrice
}

func NewBinanceSourceProvider() *BinanceSourceProvider {
	return &BinanceSourceProvider{
		symbols:    make(map[string]*sourceProvider.Symbol),
		symbolPriceData: make(map[string]*sourceProvider.SymbolPrice),
	}
}

func (b *BinanceSourceProvider) Symbols() map[string]*sourceProvider.Symbol {
	return b.symbols
}

func (b *BinanceSourceProvider) SymbolPriceData() map[string]*sourceProvider.SymbolPrice {
	return b.symbolPriceData
}

func (b *BinanceSourceProvider) GetSymbols(force bool) ([]*sourceProvider.Symbol, error) {
	// get a list of all symbols on Binance & save to file as cache
	if !force && fileHelper.PathExists(BinanceTokenListPath) {
		var symbols []*sourceProvider.Symbol
		err := jsonHelper.ReadJSONFile(BinanceTokenListPath, &symbols)

		return symbols, err
	}

	var data *map[string]interface{}
	data, err := ioHelper.Get(BinanceApiUrl+"/exchangeInfo", data)
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
			}
		}
		// save to file
		jsonHelper.WriteJSONFile(BinanceTokenListPath, symbols)
	}

	return dataMap, err
}

func (b *BinanceSourceProvider) SubscribeSymbols(symbols []*sourceProvider.Symbol) {
	// subscribe a new data stream for a new symbol
	// check if symbol already exists
	for _, symbol := range symbols {
		if _, ok := b.symbols[symbol.Symbol]; ok {
			continue
		}

		b.symbols[symbol.Symbol] = symbol
	}
	b.stopDataStream()
	b.startDataStream()
}

func (b *BinanceSourceProvider) startDataStream() {
	// subscribe to multiple data streams using one connection (ticker topic)
	// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#individual-symbol-ticker-streams
	var symbolString string
	var charCount int = 0

	for key := range b.symbols {
		symbolString += strings.ToLower(key) + "@ticker/"
	}

	if charCount = utf8.RuneCountInString(symbolString); charCount == 0 {
		return
	}

	symbolString = string([]rune(symbolString)[:charCount-1])
	var endpoint string = BinanceWsUrl + "/stream?streams=" + symbolString
	b.stream = ioHelper.NewWebSocketClient(endpoint)
	b.stream.Start(b.handleDataStream)
}

func (b *BinanceSourceProvider) handleDataStream(data *[]byte) {
	// handle the ticker data stream
	var ticker BinanceSymbolTicker
	jsonHelper.Unmarshal(*data, &ticker)
	bestAsk, _ := strconv.ParseFloat(ticker.Data.BestAskPrice, 64)
	bestBid, _ := strconv.ParseFloat(ticker.Data.BestBidPrice, 64)
	bestAskQuantity, _ := strconv.ParseFloat(ticker.Data.BestAskQuantity, 64)
	bestBidQuantity, _ := strconv.ParseFloat(ticker.Data.BestBidQuantity, 64)

	b.symbolPriceData[ticker.Data.Symbol] = &sourceProvider.SymbolPrice{
		Symbol:          b.symbols[ticker.Data.Symbol],
		BestBid:         bestBid,
		BestBidQuantity: bestBidQuantity,
		BestAsk:         bestAsk,
		BestAskQuantity: bestAskQuantity,
		EventTime:       time.Unix(0, ticker.Data.EventTime*1000000),
	}

	// build a general interface so all exchanges can use the same data structure
	// fmt.Println(string(*data))
}

func (b *BinanceSourceProvider) stopDataStream() {
	// stop the data stream if exists
	if b.stream != nil {
		b.stream.Stop()
	}
}

func (b *BinanceSourceProvider) UnsubscribeSymbol(symbol *sourceProvider.Symbol) {
	// unsubscribe a symbol from the data stream (remove symbol from the map -> restart data stream)
	delete(b.symbols, symbol.Symbol)
	b.stopDataStream()
	b.startDataStream()
}
