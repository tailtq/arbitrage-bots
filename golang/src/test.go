package main

import (
	"arbitrage-bot/services/sourceprovider"
	"arbitrage-bot/services/sourceprovider/dex"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var _sourceProvider = dex.NewPancakeswapSourceProvider()
	var symbol = sourceprovider.Symbol{
		Symbol:             "BUSDWBNB",
		BaseAsset:          "BUSD",
		BaseAssetAddress:   "0xe9e7cea3dedca5984780bafc599bd69add087d56",
		BaseAssetDecimals:  18,
		QuoteAsset:         "WBNB",
		QuoteAssetAddress:  "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
		QuoteAssetDecimals: 18,
	}
	var result = _sourceProvider.Web3Service().GetPrice(symbol, 1, "baseToQuote", true)
	fmt.Println(result)
}
