package CEX

import (
	ioHelper "arbitrage-bot/helpers/io"
	"arbitrage-bot/sourceProvider"
	"sync"
)

type OKXSourceProvider struct {
	// data stream
	streamTicker *ioHelper.WebSocketClient
    streamOrderbookDepth *ioHelper.WebSocketClient
	symbols    map[string]*sourceProvider.Symbol
    // we'll get (fatal error: concurrent map read and map write) if using regular map
	symbolPriceData sync.Map
    symbolOrderbookData sync.Map
}

func NewOKXSourceProvider() *OKXSourceProvider {
	return &OKXSourceProvider{
		symbols:    make(map[string]*sourceProvider.Symbol),
	}
}
