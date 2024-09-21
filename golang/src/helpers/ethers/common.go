package ethers

import (
	"arbitrage-bot/services/sourceprovider"
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
func GetTradePaths(symbols []sourceprovider.Symbol, tradeDirections []string) []sourceprovider.TradePath {
	var tradePaths = make([]sourceprovider.TradePath, len(symbols))

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
		tradePaths[i] = sourceprovider.TradePath{
			BaseAssetAddress:   inputTokenA,
			BaseAssetDecimals:  inputDecimalsA,
			QuoteAssetAddress:  inputTokenB,
			QuoteAssetDecimals: inputDecimalsB,
		}
	}

	return tradePaths
}

func GetPancakeSwapAddresses(networkName string) map[string]string {
	if networkName == "bsc" {
		return map[string]string{
			"factory": "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73",
			"router":  "0x10ED43C718714eb63d5aA57B78B54704E256024E",
		}
	}
	return map[string]string{}
}
