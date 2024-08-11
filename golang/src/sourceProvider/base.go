package sourceprovider

import (
	"time"
)

// ISourceProvider ... Interface for the source provider
type ISourceProvider interface {
	GetTokenListCachePath() string
	GetArbitragePairCachePath() string
	GetSymbolPrice(symbol string) *SymbolPrice
	GetSymbolOrderbookDepth(symbol string) *SymbolOrderbookDepth
	GetSymbols(force bool) ([]*Symbol, error)
	SubscribeSymbols(symbols []*Symbol)
}

// Symbol ... Represents a symbol
type Symbol struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}

// https://www.investopedia.com/terms/b/bid-and-ask.asp
// bid: the highest price a buyer will pay for a security
// ask: the lowest price a seller will take for it
// The difference between bid and ask prices, or the spread, is a key indicator of the liquidity of the asset.

// SymbolPrice ... Represents the price of a symbol
type SymbolPrice struct {
	Symbol          *Symbol   `json:"symbol"`
	BestBid         float64   `json:"bestBid"`
	BestBidQuantity float64   `json:"bestBidQty"`
	BestAsk         float64   `json:"bestAsk"`
	BestAskQuantity float64   `json:"bestAskQty"`
	EventTime       time.Time `json:"eventTime"`
}

// OrderbookEntry ... Represents an orderbook entry
type OrderbookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// SymbolOrderbookDepth ... Represents the orderbook depth of a symbol
type SymbolOrderbookDepth struct {
	Symbol       *Symbol           `json:"symbol"`
	LastUpdateID int               `json:"lastUpdateId"`
	Bids         []*OrderbookEntry `json:"bids"`
	Asks         []*OrderbookEntry `json:"asks"`
}

// SourceProviderName ... Source provider name
var SourceProviderName = map[string]string{
	"Binance": "Binance",
	"MEXC":    "MEXC",
	"Uniswap": "Uniswap",
}
