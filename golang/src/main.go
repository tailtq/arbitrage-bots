package main

import (
	"arbitrage-bot/helpers"
	fileHelper "arbitrage-bot/helpers/file"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/arbitrage"
	"arbitrage-bot/sourceProvider"
	CEX "arbitrage-bot/sourceProvider/cex"
	"fmt"
	"time"
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
    var triangularPairBatches [][3]*sourceProvider.Symbol = step1(true)
    var symbols []*sourceProvider.Symbol

    for _, pair := range triangularPairBatches {
        for _, symbol := range pair {
            symbols = append(symbols, symbol)
        }
    }

    binanceSourceProvider := CEX.NewBinanceSourceProvider()
    binanceSourceProvider.SubscribeSymbols(symbols)
    arbitrageCalculator := arbitrage.NewArbitrageCalculator(binanceSourceProvider)

    fmt.Println("Subscribed to symbols, waiting for data...")
    time.Sleep(20 * time.Second)
    fmt.Println("Starting the arbitrage calculation...")

    for {
        for _, triangularPairs := range triangularPairBatches {
            result := arbitrageCalculator.CalcTriangularArbSurfaceRate(triangularPairs)

            if result != nil && result.ProfitLoss > 0 {
                fmt.Println(result)
            }
        }
        time.Sleep(3 * time.Second)
        fmt.Println("------")
    }
    
    // symbol1 := &sourceProvider.Symbol{
    //     Symbol: "PEPEUSDT",
    //     BaseAsset: "PEPE",
    //     QuoteAsset: "USDT",
    // }
    // symbol2 := &sourceProvider.Symbol{
    //     Symbol: "BTCUSDT",
    //     BaseAsset: "BTC",
    //     QuoteAsset: "USDT",
    // }
    // symbol3 := &sourceProvider.Symbol{
    //     Symbol: "ETHUSDT",
    //     BaseAsset: "ETH",
    //     QuoteAsset: "USDT",
    // }
    // symbol4 := &sourceProvider.Symbol{
    //     Symbol: "ADAUSDT",
    //     BaseAsset: "ADA",
    //     QuoteAsset: "USDT",
    // }
    // symbol5 := &sourceProvider.Symbol{
    //     Symbol: "BNBUSDT",
    //     BaseAsset: "BNB",
    //     QuoteAsset: "USDT",
    // }
    // symbol6 := &sourceProvider.Symbol{
    //     Symbol: "SOLUSDT",
    //     BaseAsset: "SOL",
    //     QuoteAsset: "USDT",
    // }
    // symbol7 := &sourceProvider.Symbol{
    //     Symbol: "TRXUSDT",
    //     BaseAsset: "TRX",
    //     QuoteAsset: "USDT",
    // }
    // symbol8 := &sourceProvider.Symbol{
    //     Symbol: "SHIBUSDT",
    //     BaseAsset: "SHIB",
    //     QuoteAsset: "USDT",
    // }
    // var sec int = 0

    // for {
    //     time.Sleep(1 * time.Second)
    //     sec++
    //     fmt.Print("\r", sec)

    //     // if sec > 20 {
    //         // break
    //     // }
    // }

    // binanceSourceProvider.UnsubscribeSymbol(symbol1)

    // for {}
}
