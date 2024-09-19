package sourceprovider

// ISourceProvider ... Interface for the source provider
type ISourceProvider interface {
	GetArbitragePairCachePath() string
	//GetSymbols(force bool) ([]*Symbol, error)
}

// https://www.investopedia.com/terms/b/bid-and-ask.asp
// bid: the highest price a buyer will pay for a security
// ask: the lowest price a seller will take for it
// The difference between bid and ask prices, or the spread, is a key indicator of the liquidity of the asset.

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
