package sourceprovider

import "github.com/ethereum/go-ethereum/common"

// Symbol ... Represents a symbol
type Symbol struct {
	Address            string `json:"address"`
	Symbol             string `json:"symbol"`
	FeeTier            int    `json:"feeTier"` // Only used in Uniswap V3
	BaseAsset          string `json:"baseAsset"`
	BaseAssetAddress   string `json:"baseAssetAddress"`
	BaseAssetDecimals  int    `json:"baseAssetDecimals"`
	QuoteAsset         string `json:"quoteAsset"`
	QuoteAssetAddress  string `json:"quoteAssetAddress"`
	QuoteAssetDecimals int    `json:"quoteAssetDecimals"`
}

type TradePath struct {
	BaseAssetAddress   common.Address
	BaseAssetDecimals  int
	QuoteAssetAddress  common.Address
	QuoteAssetDecimals int
}
