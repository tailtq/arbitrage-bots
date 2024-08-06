package CEX

import "strings"

const BinanceApiUrl string = "https://api.binance.com/api/v3"
const BinanceWsUrl string = "wss://stream.binance.com:443"
const BinanceTokenListPath string = "data/binanceTokens.json"
const BinanceArbitragePairPath string = "data/binanceArbitragePairs.json"

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

const MEXCApiUrl string = "https://api.mexc.com/api/v3"
const MEXCWsUrl string = "wss://wbs.mexc.com/ws"
const MEXCTokenListPath string = "data/mexcTokens.json"
const MEXCArbitragePairPath string = "data/mexcArbitragePairs.json"

type MEXCEventSubscriptionUnsubscription struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

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
