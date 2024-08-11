package cex

import "arbitrage-bot/sourceProvider"

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) sourceProvider.ISourceProvider {
	if name == sourceProvider.SourceProviderName["Binance"] {
		return NewBinanceSourceProvider()
	} else if name == sourceProvider.SourceProviderName["MEXC"] {
		return NewMEXCSourceProvider()
	}
	panic("Source Provider not found")
}
