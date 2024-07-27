package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceProvider"
	CEX "arbitrage-bot/sourceProvider/cex"
	"time"
	_ "time"
)

func step1(force bool) [][3]*sourceProvider.Symbol {
    // get all the triangular pairs
    if !force && fileHelper.PathExists(CEX.BinanceArbitragePairPath) {
        var symbols [][3]*sourceProvider.Symbol
        err := jsonHelper.ReadJSONFile(CEX.BinanceArbitragePairPath, &symbols)
        helpers.Panic(err)

        return symbols
    }

    // NOTE: this doesn't cover the case when we have multiple CEX
    triangularPairFinder := arbitrage.TriangularPairFinder{}
    binanceSP := CEX.BinanceSourceProvider{}
    symbols, err := binanceSP.GetSymbols(force)
    helpers.Panic(err)

    // find the arbitrage pairs -> cache it
    triangularPairs := triangularPairFinder.Handle(symbols, 1000)
    err = jsonHelper.WriteJSONFile(CEX.BinanceArbitragePairPath, triangularPairs)
    helpers.Panic(err)

    return triangularPairs
}

func main() {
    // step1(true)

    symbol1 := &sourceProvider.Symbol{
        Symbol: "PEPEUSDT",
        BaseAsset: "PEPE",
        QuoteAsset: "USDT",
    }
    symbol2 := &sourceProvider.Symbol{
        Symbol: "BTCUSDT",
        BaseAsset: "BTC",
        QuoteAsset: "USDT",
    }
    symbol3 := &sourceProvider.Symbol{
        Symbol: "ETHUSDT",
        BaseAsset: "ETH",
        QuoteAsset: "USDT",
    }
    binanceSourceProvider := CEX.NewBinanceSourceProvider()
    binanceSourceProvider.SubscribeSymbol(symbol1)
    binanceSourceProvider.SubscribeSymbol(symbol2)
    binanceSourceProvider.SubscribeSymbol(symbol3)
    var sec int = 0

    for {
        time.Sleep(1 * time.Second)
        sec++

        // if sec > 20 {
            // break
        // }
    }

    // binanceSourceProvider.UnsubscribeSymbol(symbol1)

    // for {}
}
