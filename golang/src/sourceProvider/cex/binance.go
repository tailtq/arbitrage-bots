package CEX

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
)

type BinanceKLineCandlestick struct {
    EventType string `json:"e"`
    EventTime int64 `json:"E"`
    Symbol string `json:"s"`
    KLineData struct {
        StartTime int64 `json:"t"`
        CloseTime int64 `json:"T"`
        Symbol string `json:"s"`
        Interval string `json:"i"`
        FirstTradeID int `json:"f"`
        LastTradeID int `json:"L"`
        OpenPrice string `json:"o"`
        ClosePrice string `json:"c"`
        HighPrice string `json:"h"`
        LowPrice string `json:"l"`
        BaseAssetVolume string `json:"v"`
        NumberOfTrades int `json:"n"`
        IsClosed bool `json:"x"`
        QuoteAssetVolume string `json:"q"`
        TakerBuyBaseAssetVolume string `json:"V"`
        TakerBuyQuoteAssetVolume string `json:"Q"`
        Ignore int `json:"B"`
    } `json:"k"`
}

type BinanceSourceProvider struct {
    // data stream
    stream *ioHelper.WebSocketClient
    symbols map[string]Symbol
    symbolData map[string]SymbolPrice
}

func NewBinanceSourceProvider() *BinanceSourceProvider {
    return &BinanceSourceProvider{
        symbols: make(map[string]Symbol),
        symbolData: make(map[string]SymbolPrice),
    }
}

func (b *BinanceSourceProvider) GetTokenList(force bool) (*[]Symbol, error) {
    // get a list of all tokens on Binance & save to file as cache
    if !force && fileHelper.PathExists(BinanceCachePath) {
        var symbols []Symbol
        symbolsJSON, _ := os.ReadFile(BinanceCachePath)
        err := jsonHelper.Unmarshal(symbolsJSON, &symbols)
        
        return &symbols, err
    }

    var data *map[string]interface{}
    data, err := ioHelper.Get(BinanceApiUrl + "/exchangeInfo", data)
    helpers.Panic(err)

    dataMap := make([]Symbol, 0)
    // Type assertion (a way to retrieve the dynamic type of an interface)
    symbols, ok := (*data)["symbols"].([]interface{})
    
    if ok {
        for _, symbol := range symbols {
            if s, ok := symbol.(map[string]interface{}); ok {
                dataMap = append(dataMap, Symbol{
                    Symbol:     s["symbol"].(string),
                    BaseAsset:  s["baseAsset"].(string),
                    QuoteAsset: s["quoteAsset"].(string),
                })
            }
        }
        // save to file
        tokenListJson, _ := json.MarshalIndent(symbols, "", "\t")
        err = os.WriteFile(BinanceCachePath, tokenListJson, 0644)
    }

    return &dataMap, err
}

func (b *BinanceSourceProvider) SubscribeSymbol(symbol Symbol) {
    // subscribe a new data stream for a new symbol
    // check if symbol already exists
    if _, ok := b.symbols[symbol.Symbol]; ok {
        return
    }

    b.symbols[symbol.Symbol] = symbol
    b.stopDataStream()
    b.startDataStream()
}

func (b *BinanceSourceProvider) startDataStream() {
    // subscribe to multiple data streams using one connection
    var symbolStringString string
    var charCount int = 0

    for key := range b.symbols {
        symbolStringString += strings.ToLower(key) +  "@kline_1m/"
    }

    if charCount := utf8.RuneCountInString(symbolStringString); charCount == 0 {
        return
    }

    symbolStringString = string([]rune(symbolStringString)[:charCount - 1])
    var endpoint string = "wss://fstream.binance.com/ws/" + symbolStringString
    b.stream = ioHelper.NewWebSocketClient(endpoint)
    b.stream.Start(b.handleDataStream)
}

func (b *BinanceSourceProvider) handleDataStream(data *[]byte) {
    var kline BinanceKLineCandlestick
    jsonHelper.Unmarshal(*data, &kline)

    price, err := strconv.ParseFloat(kline.KLineData.ClosePrice, 64)
    helpers.Panic(err)

    b.symbolData[kline.Symbol] = SymbolPrice{
        Symbol: b.symbols[kline.Symbol],
        Price: price,
        EventTime: time.Unix(0, kline.EventTime * 1000000),
    }

    // build a general interface so all exchanges can use the same data structure
    fmt.Println(
        kline.Symbol,
        kline.KLineData.OpenPrice,
        kline.KLineData.ClosePrice,
        kline.KLineData.CloseTime,
        kline.KLineData.IsClosed,
        b.symbolData,
        "ehehehe",
    )
}

func (b *BinanceSourceProvider) stopDataStream() {
    // stop the data stream if exists
    if b.stream != nil {
        b.stream.Stop()
    }
}

func (b *BinanceSourceProvider) UnsubscribeSymbol(symbol Symbol) {
    // unsubscribe a symbol from the data stream (remove symbol from the map -> restart data stream)
    delete(b.symbols, symbol.Symbol)
    b.stopDataStream()
    b.startDataStream()
}
