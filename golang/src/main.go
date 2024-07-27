package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceProvider"
	CEX "arbitrage-bot/sourceProvider/cex"
	"fmt"
	_ "time"
)

func step1(force bool) [][3]*sourceProvider.Symbol {
    if !force && fileHelper.PathExists(CEX.BinanceArbitragePairPath) {
        var tokens [][3]*sourceProvider.Symbol
        err := jsonHelper.ReadJSONFile(CEX.BinanceArbitragePairPath, &tokens)
        helpers.Panic(err)

        return tokens
    }

    // NOTE: this doesn't cover the case when we have multiple CEX
    arbitragePairFinder := arbitrage.ArbitragePairFinder{}
    binanceSP := CEX.BinanceSourceProvider{}
    tokenList, err := binanceSP.GetTokenList(force)
    helpers.Panic(err)

    // find the arbitrage pairs -> cache it
    arbitragePairs := arbitragePairFinder.Handle(tokenList, 1000)
    err = jsonHelper.WriteJSONFile(CEX.BinanceArbitragePairPath, arbitragePairs)
    helpers.Panic(err)

    return arbitragePairs
}

func main() {
    step1(true)
    // fmt.Println(tokenList)

    // token1 := &sourceProvider.Symbol{
    //     Symbol: "STRKUSDT",
    //     BaseAsset: "STRK",
    //     QuoteAsset: "USDT",
    // }
    // token2 := &sourceProvider.Symbol{
    //     Symbol: "BTCUSDT",
    //     BaseAsset: "BTC",
    //     QuoteAsset: "USDT",
    // }
    // binanceSourceProvider := CEX.NewBinanceSourceProvider()
    // binanceSourceProvider.SubscribeSymbol(token1)
    // binanceSourceProvider.SubscribeSymbol(token2)
    // var sec int = 0

    // for {
    //     time.Sleep(1 * time.Second)
    //     sec++

    //     if sec > 20 {
    //         break
    //     }
    // }

    // binanceSourceProvider.UnsubscribeSymbol(token1)

    // for {}
}
