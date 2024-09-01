package dex

import (
	"arbitrage-bot/services/sourceprovider"
)

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) ISourceProvider {
	if name == sourceprovider.SourceProviderName["Uniswap"] {
		return NewUniswapSourceProviderService()
	}
	panic("Source Provider not found")
}
