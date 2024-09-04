package sourceprovider

// Symbol ... Represents a symbol
type Symbol struct {
	Address            string `json:"address"`
	Symbol             string `json:"symbol"`
	FeeTier            int    `json:"feeTier"`
	BaseAsset          string `json:"baseAsset"`
	BaseAssetAddress   string `json:"baseAssetAddress"`
	BaseAssetDecimals  int    `json:"baseAssetDecimals"`
	QuoteAsset         string `json:"quoteAsset"`
	QuoteAssetAddress  string `json:"quoteAssetAddress"`
	QuoteAssetDecimals int    `json:"quoteAssetDecimals"`
}
