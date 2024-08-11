package dex

import "arbitrage-bot/sourceprovider"

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) sourceprovider.ISourceProvider {
	if name == sourceprovider.SourceProviderName["Uniswap"] {
		return NewUniswapSourceProvider()
	}
	panic("Source Provider not found")
}
