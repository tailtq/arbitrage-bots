package cex

import "arbitrage-bot/sourceprovider"

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) sourceprovider.ISourceProvider {
	if name == sourceprovider.SourceProviderName["Binance"] {
		return NewBinanceSourceProvider()
	} else if name == sourceprovider.SourceProviderName["MEXC"] {
		return NewMEXCSourceProvider()
	}
	panic("Source Provider not found")
}
