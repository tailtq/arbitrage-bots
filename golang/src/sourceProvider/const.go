package sourceprovider

// Symbol ... Represents a symbol
type Symbol struct {
	Symbol             string `json:"symbol"`
	BaseAsset          string `json:"baseAsset"`
	BaseAssetID        string `json:"baseAssetId"`
	BaseAssetDecimals  uint8  `json:"baseAssetDecimals"`
	QuoteAsset         string `json:"quoteAsset"`
	QuoteAssetID       string `json:"quoteAssetId"`
	QuoteAssetDecimals uint8  `json:"quoteAssetDecimals"`
}
