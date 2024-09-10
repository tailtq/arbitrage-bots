package web3

import (
	"arbitrage-bot/helpers"
	"arbitrage-bot/helpers/ethers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/sourceprovider"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"
	"sync"
)

type PancakeswapWeb3Service struct {
	client          *ethclient.Client
	factoryContract *bind.BoundContract
	routerContract  *bind.BoundContract
}

func NewPancakeswapWeb3Service() *PancakeswapWeb3Service {
	var networkRpcUrl = os.Getenv("NETWORK_RPC_URL")
	var networkName = os.Getenv("NETWORK_NAME")
	var pancakeswapAddresses = GetPancakeSwapAddresses(networkName)
	var factoryAddress = common.HexToAddress(pancakeswapAddresses["factory"])
	var routerAddress = common.HexToAddress(pancakeswapAddresses["router"])
	factoryABI, err := jsonHelper.ReadJSONABIFile("data/web3/pancakeswapFactoryV2ABI.json")
	helpers.Panic(err)
	routerABI, err := jsonHelper.ReadJSONABIFile("data/web3/pancakeswapRouterABI.json")
	helpers.Panic(err)
	client, err := ethclient.Dial(networkRpcUrl)
	helpers.Panic(err)

	var factoryContract = bind.NewBoundContract(factoryAddress, factoryABI, client, client, client)
	var routerContract = bind.NewBoundContract(routerAddress, routerABI, client, client, client)

	return &PancakeswapWeb3Service{
		client:          client,
		factoryContract: factoryContract,
		routerContract:  routerContract,
	}
}

func (u *PancakeswapWeb3Service) GetPrice(symbol sourceprovider.Symbol, amountIn float64, tradeDirection string, verbose bool) float64 {
	if amountIn == 0 {
		return 0
	}
	var inputTokenA, inputTokenB common.Address
	var inputDecimalsA, inputDecimalsB int

	if tradeDirection == "baseToQuote" {
		inputTokenA = common.HexToAddress(symbol.BaseAssetAddress)
		inputDecimalsA = symbol.BaseAssetDecimals
		inputTokenB = common.HexToAddress(symbol.QuoteAssetAddress)
		inputDecimalsB = symbol.QuoteAssetDecimals
	} else if tradeDirection == "quoteToBase" {
		inputTokenA = common.HexToAddress(symbol.QuoteAssetAddress)
		inputDecimalsA = symbol.QuoteAssetDecimals
		inputTokenB = common.HexToAddress(symbol.BaseAssetAddress)
		inputDecimalsB = symbol.BaseAssetDecimals
	}

	var amountInParsed = ethers.EtherToWei(amountIn, inputDecimalsA)
	var result []interface{}
	var tradePath = []common.Address{inputTokenA, inputTokenB}
	var err = u.routerContract.Call(&bind.CallOpts{}, &result, "getAmountsOut", amountInParsed, tradePath)

	if err != nil {
		if verbose {
			fmt.Println("Router error:", symbol.Symbol, err)
		}
		return 0
	}

	var priceList = result[0].([]*big.Int)
	return ethers.WeiToEther(priceList[len(priceList)-1], inputDecimalsB)
}

func (u *PancakeswapWeb3Service) AggregatePrices(symbols []*sourceprovider.Symbol, verbose bool) *sync.Map {
	var channel = make(chan *sourceprovider.Symbol)
	var concurrency = 8
	var result sync.Map
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for range concurrency {
		go func() {
			defer wg.Done()
			for symbol := range channel {
				var price = u.GetPrice(*symbol, 1, "baseToQuote", verbose)
				if price != 0 {
					result.Store(symbol.Symbol, price)
				}
			}
		}()
	}
	for _, symbol := range symbols {
		channel <- symbol
	}
	close(channel)
	wg.Wait()

	return &result
}
