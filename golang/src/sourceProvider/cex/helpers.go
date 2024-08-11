package cex

import "arbitrage-bot/sourceProvider"

// GetCEXSourceProvider ... Get the source provider for the CEX
func GetCEXSourceProvider(name string) sourceProvider.ISourceProvider {
	if name == SourceProviderName["Binance"] {
		return NewBinanceSourceProvider()
	} else if name == SourceProviderName["MEXC"] {
		return NewMEXCSourceProvider()
	}
	panic("Source Provider not found")
}
