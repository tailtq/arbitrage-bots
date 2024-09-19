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

// NewPancakeswapWeb3Service ... creates a new PancakeswapWeb3Service
func NewPancakeswapWeb3Service() *PancakeswapWeb3Service {
	var networkRpcUrl = os.Getenv("NETWORK_RPC_URL")
	var networkName = os.Getenv("NETWORK_NAME")
	var pancakeswapAddresses = ethers.GetPancakeSwapAddresses(networkName)
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

// GetPrice ... gets price
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

// AggregatePrices ... aggregates prices
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

// GetPoolDataByIndex ... gets pool data by index
func (u *PancakeswapWeb3Service) GetPoolDataByIndex(index int) sourceprovider.Symbol {
	var result []interface{}
	var err = u.factoryContract.Call(&bind.CallOpts{}, &result, "allPairs", big.NewInt(int64(index)))
	helpers.Panic(err)
	var poolAddress = result[0].(common.Address)

	return u.GetPoolData(poolAddress)
}

func (u *PancakeswapWeb3Service) GetPoolData(address common.Address) sourceprovider.Symbol {
	poolABI, err := jsonHelper.ReadJSONABIFile("data/web3/pancakeswapPoolABI.json")
	helpers.Panic(err)
	erc20ABI, err := jsonHelper.ReadJSONABIFile("data/web3/erc20.json")
	helpers.Panic(err)

	var poolContract = bind.NewBoundContract(address, poolABI, u.client, u.client, u.client)
	var wg = sync.WaitGroup{}
	wg.Add(2)
	var symbol sourceprovider.Symbol
	var resultToken0, resultToken1 []interface{}
	var errToken0, errToken1 error
	go ethers.CallContractMethod(&wg, poolContract, "token0", []interface{}{}, &resultToken0, &errToken0)
	go ethers.CallContractMethod(&wg, poolContract, "token1", []interface{}{}, &resultToken1, &errToken1)
	//go ethers.CallContractMethod(&wg, poolContract, "fee", []interface{}{}, &resultFee, &errFee)
	wg.Wait()
	helpers.PanicBatch(errToken0, errToken1)

	for _, result := range [][]interface{}{resultToken0, resultToken1} {
		wg.Add(2)
		var tokenAddress = result[0].(common.Address)
		var tokenContract = bind.NewBoundContract(tokenAddress, erc20ABI, u.client, u.client, u.client)
		var resultSymbol, resultDecimals []interface{}
		var errSymbol, errDecimals error
		go ethers.CallContractMethod(&wg, tokenContract, "symbol", []interface{}{}, &resultSymbol, &errSymbol)
		go ethers.CallContractMethod(&wg, tokenContract, "decimals", []interface{}{}, &resultDecimals, &errDecimals)
		wg.Wait()
		helpers.PanicBatch(errSymbol, errDecimals)

		if tokenAddress == resultToken0[0] {
			symbol.BaseAsset = resultSymbol[0].(string)
			symbol.BaseAssetAddress = tokenAddress.String()
			symbol.BaseAssetDecimals = int(resultDecimals[0].(uint8))
		} else {
			symbol.QuoteAsset = resultSymbol[0].(string)
			symbol.QuoteAssetAddress = tokenAddress.String()
			symbol.QuoteAssetDecimals = int(resultDecimals[0].(uint8))
		}
	}
	symbol.Address = address.String()
	symbol.Symbol = symbol.BaseAsset + symbol.QuoteAsset
	//symbol.FeeTier = int(resultFee[0].(*big.Int).Int64())

	return symbol
}
