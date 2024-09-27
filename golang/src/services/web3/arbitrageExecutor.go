package web3

import (
	"arbitrage-bot/helpers"
	ethersHelper "arbitrage-bot/helpers/ethers"
	jsonHelper "arbitrage-bot/helpers/json"
	sp "arbitrage-bot/services/sourceprovider"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math"
	"math/big"
	"os"
	"slices"
)

type SwapParams struct {
	Protocol uint8          `json:"protocol"`
	TokenIn  common.Address `json:"tokenIn"`
	TokenOut common.Address `json:"tokenOut"`
	Fee      *big.Int       `json:"fee"`
}

type ArbitrageExecutorWeb3Service struct {
	client          *ethclient.Client
	contract        *bind.BoundContract
	contractABI     abi.ABI
	contractAddress common.Address
}

func NewArbitrageExecutorWeb3Service() *ArbitrageExecutorWeb3Service {
	var networkRpcUrl = os.Getenv("NETWORK_RPC_URL")
	var networkName = os.Getenv("NETWORK_NAME")
	var contractAddress = ethersHelper.GetArbitrageExecutorAddresses(networkName)["v1"]

	client, err := ethclient.Dial(networkRpcUrl)
	helpers.Panic(err)
	contractABI, err := jsonHelper.ReadJSONABIFile("data/web3/arbitrageExecutorABI.json")
	helpers.Panic(err)
	var contract = bind.NewBoundContract(contractAddress, contractABI, client, client, client)

	return &ArbitrageExecutorWeb3Service{
		client:          client,
		contract:        contract,
		contractAddress: contractAddress,
		contractABI:     contractABI,
	}
}

func (a *ArbitrageExecutorWeb3Service) ExecuteArbitrage(
	tradePaths []sp.TradePath,
	amountIn float64,
	loanAddress common.Address,
) error {
	var amountInParsed = big.NewInt(int64(amountIn * math.Pow(10, float64(tradePaths[0].BaseAssetDecimals))))
	var swapParams []SwapParams

	for _, tradePath := range tradePaths {
		swapParams = append(swapParams, SwapParams{
			Protocol: 0,
			TokenIn:  tradePath.BaseAssetAddress,
			TokenOut: tradePath.QuoteAssetAddress,
			Fee:      big.NewInt(0),
		})
	}

	data, err := a.contractABI.Pack("swapIn", swapParams, amountInParsed, loanAddress)
	helpers.Panic(err)
	message := ethereum.CallMsg{To: &a.contractAddress, Data: data}
	// TODO: We've got work over here, what to do with the estimated gas?
	//var startEstimatingTime = time.Now()
	//estimatedGas, err := a.client.EstimateGas(context.Background(), message)
	//fmt.Println("Estimating time:", time.Since(startEstimatingTime))
	//fmt.Println("Estimated gas:", estimatedGas)
	//helpers.Panic(err)
	_, err = a.client.CallContract(context.Background(), message, nil)

	return err
}

// GetLoanAddress ... gets other loan address as PancakeSwap or UniswapV2 don't support borrowing the token in the path
// in FlashSwap
func (a *ArbitrageExecutorWeb3Service) GetLoanAddress(symbols []*sp.Symbol, paths []sp.TradePath) common.Address {
	var pathAddresses []common.Address
	var result common.Address

	for _, path := range paths {
		pathAddresses = append(pathAddresses, path.BaseAssetAddress)
		pathAddresses = append(pathAddresses, path.QuoteAssetAddress)
	}

	for _, symbol := range symbols {
		var baseAssetAddress = common.HexToAddress(symbol.BaseAssetAddress)
		var quoteAssetAddress = common.HexToAddress(symbol.QuoteAssetAddress)

		if !slices.Contains(pathAddresses, baseAssetAddress) {
			result = baseAssetAddress
			break
		} else if !slices.Contains(pathAddresses, quoteAssetAddress) {
			result = quoteAssetAddress
			break
		}
	}

	if (result == common.Address{}) {
		helpers.Panic(fmt.Errorf("No loan address found"))
	}

	return result
}
