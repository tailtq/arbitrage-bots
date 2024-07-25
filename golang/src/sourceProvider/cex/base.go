package CEX

import "time"

type SourceProvider interface {
    GetTokenList() (*[]Symbol, error)
}

type Symbol struct {
    Symbol string `json:"symbol"`
    BaseAsset string `json:"baseAsset"`
    QuoteAsset string `json:"quoteAsset"`
}

type SymbolPrice struct {
    Symbol Symbol `json:"symbol"`
    Price float64 `json:"price"`
    EventTime time.Time `json:"eventTime"`
}

// type
