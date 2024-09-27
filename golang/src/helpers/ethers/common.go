package ethers

import (
	"arbitrage-bot/models"
	sp "arbitrage-bot/services/sourceprovider"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math"
	"math/big"
	"sync"
)

func WeiToEther(wei *big.Int, decimals int) float64 {
	f := new(big.Float).SetInt(wei)
	ethValue, _ := f.Quo(f, big.NewFloat(math.Pow10(decimals))).Float64()
	return ethValue
}

func EtherToWei(eth float64, decimals int) *big.Int {
	return big.NewInt(int64(eth * math.Pow(10, float64(decimals))))
}

// CallContractMethod ... call contract method asynchronously
func CallContractMethod(
	wg *sync.WaitGroup,
	contract *bind.BoundContract,
	methodName string,
	params []interface{},
	result *[]interface{},
	err *error,
) {
	defer wg.Done()
	*err = contract.Call(nil, result, methodName, params...)
}

// GetTradePaths ... get trade paths from symbols and trade directions
func GetTradePaths(symbols []sp.Symbol, tradeDirections []string) []sp.TradePath {
	var tradePaths = make([]sp.TradePath, len(symbols))

	for i, symbol := range symbols {
		var inputTokenA, inputTokenB common.Address
		var inputDecimalsA, inputDecimalsB int

		if tradeDirections[i] == "baseToQuote" {
			inputTokenA = common.HexToAddress(symbol.BaseAssetAddress)
			inputDecimalsA = symbol.BaseAssetDecimals
			inputTokenB = common.HexToAddress(symbol.QuoteAssetAddress)
			inputDecimalsB = symbol.QuoteAssetDecimals
		} else if tradeDirections[i] == "quoteToBase" {
			inputTokenA = common.HexToAddress(symbol.QuoteAssetAddress)
			inputDecimalsA = symbol.QuoteAssetDecimals
			inputTokenB = common.HexToAddress(symbol.BaseAssetAddress)
			inputDecimalsB = symbol.BaseAssetDecimals
		}
		tradePaths[i] = sp.TradePath{
			BaseAssetAddress:   inputTokenA,
			BaseAssetDecimals:  inputDecimalsA,
			QuoteAssetAddress:  inputTokenB,
			QuoteAssetDecimals: inputDecimalsB,
		}
	}

	return tradePaths
}

func GetTradePathsFromSurfaceResult(surfaceResult models.TriangularArbSurfaceResult) []sp.TradePath {
	var symbols = []sp.Symbol{
		surfaceResult.Symbol1,
		surfaceResult.Symbol2,
		surfaceResult.Symbol3,
	}
	var tradeDirections = []string{
		surfaceResult.DirectionTrade1,
		surfaceResult.DirectionTrade2,
		surfaceResult.DirectionTrade3,
	}

	return GetTradePaths(symbols, tradeDirections)
}

func GetPancakeSwapAddresses(networkName string) map[string]string {
	if networkName == "bsc" {
		return map[string]string{
			"factory": "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73",
			"router":  "0x10ED43C718714eb63d5aA57B78B54704E256024E",
		}
	} else if networkName == "bsc-testnet" {
		return map[string]string{
			"factory": "0xB7926C0430Afb07AA7DEfDE6DA862aE0Bde767bc",
			"router":  "0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5c3",
		}
	}
	return map[string]string{}
}

func GetArbitrageExecutorAddresses(networkName string) map[string]common.Address {
	if networkName == "bsc-testnet" {
		return map[string]common.Address{
			"v1": common.HexToAddress("0x1959b2a1776dee3daef75ae6b545f9c8d6b0df6b"),
		}
	}
	return map[string]common.Address{}
}
