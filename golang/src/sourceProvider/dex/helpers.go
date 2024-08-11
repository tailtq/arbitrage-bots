package dex

import "arbitrage-bot/sourceProvider"

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) sourceProvider.ISourceProvider {
	if name == sourceProvider.SourceProviderName["Uniswap"] {
		return NewUniswapSourceProvider()
	}
	panic("Source Provider not found")
}
