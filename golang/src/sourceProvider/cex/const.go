package CEX

const BinanceApiUrl string = "https://api.binance.com/api/v3"
const BinanceWsUrl string = "wss://stream.binance.com:443"
const BinanceTokenListPath string = "data/binanceTokens.json"
const BinanceArbitragePairPath string = "data/binanceArbitragePairs.json"

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

type BinanceSymbolTicker struct {
    Stream string `json:"stream"`
    Data struct {
        EventType string `json:"e"`
        EventTime int64 `json:"E"`
        Symbol string `json:"s"`
        PriceChange string `json:"p"`
        PriceChangePercent string `json:"P"`
        WeightedAveragePrice string `json:"w"`
        FirstTradePrice string `json:"x"`
        LastPrice string `json:"c"`
        LastQuantity string `json:"Q"`
        BestBidPrice string `json:"b"`
        BestBidQuantity string `json:"B"`
        BestAskPrice string `json:"a"`
        BestAskQuantity string `json:"A"`
        OpenPrice string `json:"o"`
        HighPrice string `json:"h"`
        LowPrice string `json:"l"`
        TotalTradedBaseAssetVolume string `json:"v"`
        TotalTradedQuoteAssetVolume string `json:"q"`
        StatisticsOpenTime int `json:"O"`
        StatisticsCloseTime int `json:"C"`
        FirstTradeID int `json:"F"`
        LastTradeID int `json:"L"`
        TotalNumberOfTrades int `json:"n"`
    } `json:"data"`
}
