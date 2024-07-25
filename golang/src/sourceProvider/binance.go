package sourceProvider

import (
	fileHelper "arbitrage-bot/helpers/file"
	ioHelper "arbitrage-bot/helpers/io"
	jsonHelper "arbitrage-bot/helpers/json"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type BinanceSourceProvider struct {
    // data stream
    stream *ioHelper.WebSocketClient
    Symbols []Symbol
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

    if err != nil {
        panic(err)
    }

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
    for _, existingSymbol := range b.Symbols {
        if existingSymbol.Symbol == symbol.Symbol {
            return
        }
    }

    b.Symbols = append(b.Symbols, symbol)
    b.stopDataStream()
    b.startDataStream()
}

func (b *BinanceSourceProvider) startDataStream() {
    // subscribe to multiple data streams using one connection
    var symbolStringString string
    var charCount int = 0

    for _, symbol := range b.Symbols {
        symbolStringString += strings.ToLower(symbol.Symbol) +  "@aggTrade/"
    }

    charCount = utf8.RuneCountInString(symbolStringString)

    if charCount == 0 {
        return
    }

    symbolStringString = string([]rune(symbolStringString)[:charCount - 1])
    var endpoint string = "wss://fstream.binance.com/ws/" + symbolStringString
    b.stream = ioHelper.NewWebSocketClient(endpoint)
    b.stream.Start(func(data interface{}) {
        fmt.Println(data)
    })
}

func (b *BinanceSourceProvider) stopDataStream() {
    if b.stream != nil {
        b.stream.Stop()
    }
}

func (b *BinanceSourceProvider) UnsubscribeSymbol(symbol Symbol) {
    // unsubscribe a symbol from the data stream (remove symbol from the list -> restart data stream)
    for i, existingSymbol := range b.Symbols {
        if existingSymbol.Symbol == symbol.Symbol {
            b.Symbols = append(b.Symbols[:i], b.Symbols[i+1:]...)
            break
        }
    }

    b.stopDataStream()
    b.startDataStream()
}
