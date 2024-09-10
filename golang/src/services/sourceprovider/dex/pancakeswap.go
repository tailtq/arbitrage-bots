package dex

import (
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/web3"
	"sync"
)

type PancakeswapSourceProvider struct {
	web3Service     *web3.PancakeswapWeb3Service
	symbolPriceData sync.Map
	symbols         map[string]*sourceprovider.Symbol
}

func NewPancakeswapSourceProvider() *PancakeswapSourceProvider {
	return &PancakeswapSourceProvider{
		web3Service: web3.NewPancakeswapWeb3Service(),
	}
}

func (u *PancakeswapSourceProvider) Web3Service() *web3.PancakeswapWeb3Service {
	return u.web3Service
}
