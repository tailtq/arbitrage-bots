package sourceProvider

import "time"

type SourceProvider interface {
    GetSymbols() (*[]Symbol, error)
}

type Symbol struct {
    Symbol string `json:"symbol"`
    BaseAsset string `json:"baseAsset"`
    QuoteAsset string `json:"quoteAsset"`
}

type SymbolPrice struct {
    Symbol *Symbol `json:"symbol"`
    BestBid float64 `json:"bestBid"`
    BestBidQuantity float64 `json:"bestBidQty"`
    BestAsk float64 `json:"bestAsk"`
    BestAskQuantity float64 `json:"bestAskQty"`
    EventTime time.Time `json:"eventTime"`
}

// type
