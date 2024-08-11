package cex

import "strings"

// BinanceAPIURL ... Binance API URL
const BinanceAPIURL string = "https://api.binance.com/api/v3"

// BinanceWsURL ... Binance WS URL
const BinanceWsURL string = "wss://stream.binance.com:443"

// BinanceTokenListPath ... Binance token list path
const BinanceTokenListPath string = "data/binanceTokens.json"

// BinanceArbitragePairPath ... Binance arbitrage pair path
const BinanceArbitragePairPath string = "data/binanceArbitragePairs.json"

// BinanceSymbolTicker ... Binance symbol ticker
type BinanceSymbolTicker struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType                   string `json:"e"`
		EventTime                   int64  `json:"E"`
		Symbol                      string `json:"s"`
		PriceChange                 string `json:"p"`
		PriceChangePercent          string `json:"P"`
		WeightedAveragePrice        string `json:"w"`
		FirstTradePrice             string `json:"x"`
		LastPrice                   string `json:"c"`
		LastQuantity                string `json:"Q"`
		BestBidPrice                string `json:"b"`
		BestBidQuantity             string `json:"B"`
		BestAskPrice                string `json:"a"`
		BestAskQuantity             string `json:"A"`
		OpenPrice                   string `json:"o"`
		HighPrice                   string `json:"h"`
		LowPrice                    string `json:"l"`
		TotalTradedBaseAssetVolume  string `json:"v"`
		TotalTradedQuoteAssetVolume string `json:"q"`
		StatisticsOpenTime          int    `json:"O"`
		StatisticsCloseTime         int    `json:"C"`
		FirstTradeID                int    `json:"F"`
		LastTradeID                 int    `json:"L"`
		TotalNumberOfTrades         int    `json:"n"`
	} `json:"data"`
}

// BinanceOrderbookDepth ... Binance orderbook depth
type BinanceOrderbookDepth struct {
	Stream string `json:"stream"`
	Data   struct {
		LastUpdateID int        `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	} `json:"data"`
}

// GetSymbol ... returns the symbol from the stream
func (b *BinanceOrderbookDepth) GetSymbol() string {
	return strings.ToUpper(strings.Split(b.Stream, "@")[0])
}

// MEXCAPIURL ... MEXC API URL
const MEXCAPIURL string = "https://api.mexc.com/api/v3"

// MEXCWsURL ... MEXC WS URL
const MEXCWsURL string = "wss://wbs.mexc.com/ws"

// MEXCTokenListPath ... MEXC token list path
const MEXCTokenListPath string = "data/mexcTokens.json"

// MEXCArbitragePairPath ... MEXC arbitrage pair path
const MEXCArbitragePairPath string = "data/mexcArbitragePairs.json"

// MEXCEventSubscriptionUnsubscription ... MEXC event subscription subscription
type MEXCEventSubscriptionUnsubscription struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

// MEXCSymbolTicker ... MEXC symbol ticker
type MEXCSymbolTicker struct {
	Channel string `json:"c"`
	Data    struct {
		BestBidPrice    string `json:"b"`
		BestBidQuantity string `json:"B"`
		BestAskPrice    string `json:"a"`
		BestAskQuantity string `json:"A"`
	} `json:"d"`
	Symbol string `json:"s"`
	Time   int64  `json:"t"`
}

// SourceProviderName ... Source provider name
var SourceProviderName = map[string]string{
	"Binance": "Binance",
	"MEXC":    "MEXC",
}
