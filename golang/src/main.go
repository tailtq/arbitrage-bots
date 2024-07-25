package main

import (
	"arbitrage-bot/helpers"
	CEX "arbitrage-bot/sourceProvider/cex"
	"time"
)

func main() {
    binanceSP := CEX.BinanceSourceProvider{}
    tokenList, err := binanceSP.GetTokenList(false)
    helpers.Panic(err)
    _ = tokenList
    // fmt.Println(tokenList)

    symbol1 := CEX.Symbol{
        Symbol: "STRKUSDT",
        BaseAsset: "STRK",
        QuoteAsset: "USDT",
    }
    symbol2 := CEX.Symbol{
        Symbol: "BTCUSDT",
        BaseAsset: "BTC",
        QuoteAsset: "USDT",
    }
    binanceSourceProvider := CEX.NewBinanceSourceProvider()
    binanceSourceProvider.SubscribeSymbol(symbol1)
    binanceSourceProvider.SubscribeSymbol(symbol2)
    var sec int = 0

    for {
        time.Sleep(1 * time.Second)
        sec++

        if sec > 20 {
            break
        }
    }

    binanceSourceProvider.UnsubscribeSymbol(symbol1)

    for {}
}
