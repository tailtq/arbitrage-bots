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
	"arbitrage-bot/sourceProvider"
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
    tokens map[string]*sourceProvider.Symbol
    tokenData map[string]*sourceProvider.SymbolPrice
}

func NewBinanceSourceProvider() *BinanceSourceProvider {
    return &BinanceSourceProvider{
        tokens: make(map[string]*sourceProvider.Symbol),
        tokenData: make(map[string]*sourceProvider.SymbolPrice),
    }
}

func (b *BinanceSourceProvider) GetTokenList(force bool) ([]*sourceProvider.Symbol, error) {
    // get a list of all tokens on Binance & save to file as cache
    if !force && fileHelper.PathExists(BinanceTokenListPath) {
        var tokens []*sourceProvider.Symbol
        err := jsonHelper.ReadJSONFile(BinanceTokenListPath, &tokens)
        
        return tokens, err
    }

    var data *map[string]interface{}
    data, err := ioHelper.Get(BinanceApiUrl + "/exchangeInfo", data)
    helpers.Panic(err)

    dataMap := make([]*sourceProvider.Symbol, 0)
    // Type assertion (a way to retrieve the dynamic type of an interface)
    tokens, ok := (*data)["symbols"].([]interface{})
    
    if ok {
        for _, token := range tokens {
            if s, ok := token.(map[string]interface{}); ok && s["isSpotTradingAllowed"].(bool) {
                dataMap = append(dataMap, &sourceProvider.Symbol{
                    Symbol:     s["symbol"].(string),
                    BaseAsset:  s["baseAsset"].(string),
                    QuoteAsset: s["quoteAsset"].(string),
                })
            }
        }
        // save to file
        tokenListJson, _ := json.MarshalIndent(tokens, "", "\t")
        err = os.WriteFile(BinanceTokenListPath, tokenListJson, 0644)
    }

    return dataMap, err
}

func (b *BinanceSourceProvider) SubscribeSymbol(token *sourceProvider.Symbol) {
    // subscribe a new data stream for a new token
    // check if token already exists
    if _, ok := b.tokens[token.Symbol]; ok {
        return
    }

    b.tokens[token.Symbol] = token
    b.stopDataStream()
    b.startDataStream()
}

func (b *BinanceSourceProvider) startDataStream() {
    // subscribe to multiple data streams using one connection
    var tokenString string
    var charCount int = 0

    for key := range b.tokens {
        tokenString += strings.ToLower(key) +  "@kline_1m/"
    }

    if charCount := utf8.RuneCountInString(tokenString); charCount == 0 {
        return
    }

    tokenString = string([]rune(tokenString)[:charCount - 1])
    var endpoint string = "wss://fstream.binance.com/ws/" + tokenString
    b.stream = ioHelper.NewWebSocketClient(endpoint)
    b.stream.Start(b.handleDataStream)
}

func (b *BinanceSourceProvider) handleDataStream(data *[]byte) {
    var kline BinanceKLineCandlestick
    jsonHelper.Unmarshal(*data, &kline)

    price, err := strconv.ParseFloat(kline.KLineData.ClosePrice, 64)
    helpers.Panic(err)

    b.tokenData[kline.Symbol] = &sourceProvider.SymbolPrice{
        Symbol: b.tokens[kline.Symbol],
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
        b.tokenData,
        "ehehehe",
    )
}

func (b *BinanceSourceProvider) stopDataStream() {
    // stop the data stream if exists
    if b.stream != nil {
        b.stream.Stop()
    }
}

func (b *BinanceSourceProvider) UnsubscribeSymbol(token *sourceProvider.Symbol) {
    // unsubscribe a token from the data stream (remove token from the map -> restart data stream)
    delete(b.tokens, token.Symbol)
    b.stopDataStream()
    b.startDataStream()
}
