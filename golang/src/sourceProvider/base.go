package sourceProvider

import "time"

type SourceProvider interface {
    Symbols() map[string]*Symbol
    GetSymbolPrice(symbol string) *SymbolPrice
    GetSymbols(force bool) ([]*Symbol, error)
}

type Symbol struct {
    Symbol string `json:"symbol"`
    BaseAsset string `json:"baseAsset"`
    QuoteAsset string `json:"quoteAsset"`
}

// https://www.investopedia.com/terms/b/bid-and-ask.asp
// bid: the highest price a buyer will pay for a security
// ask: the lowest price a seller will take for it
// The difference between bid and ask prices, or the spread, is a key indicator of the liquidity of the asset. 
type SymbolPrice struct {
    Symbol *Symbol `json:"symbol"`
    BestBid float64 `json:"bestBid"`
    BestBidQuantity float64 `json:"bestBidQty"`
    BestAsk float64 `json:"bestAsk"`
    BestAskQuantity float64 `json:"bestAskQty"`
    EventTime time.Time `json:"eventTime"`
}
