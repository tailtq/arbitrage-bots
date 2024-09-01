package cex

import (
	"arbitrage-bot/services/sourceprovider"
)

// GetSourceProvider ... Get the source provider for the CEX
func GetSourceProvider(name string) ISourceProvider {
	if name == sourceprovider.SourceProviderName["Binance"] {
		return NewBinanceSourceProviderService()
	} else if name == sourceprovider.SourceProviderName["MEXC"] {
		return NewMEXCSourceProviderService()
	}
	panic("Source Provider not found")
}
