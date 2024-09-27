package web3

import (
	"arbitrage-bot/helpers"
	ethersHelper "arbitrage-bot/helpers/ethers"
	jsonHelper "arbitrage-bot/helpers/json"
	sp "arbitrage-bot/services/sourceprovider"
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
	var pancakeswapAddresses = ethersHelper.GetPancakeSwapAddresses(networkName)
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
func (u *PancakeswapWeb3Service) GetPrice(
	symbol sp.Symbol,
	amountIn float64,
	tradeDirection string,
	verbose bool,
) float64 {
	var tradePath = ethersHelper.GetTradePaths([]sp.Symbol{symbol}, []string{tradeDirection})[0]
	var amountInParsed = ethersHelper.EtherToWei(amountIn, tradePath.BaseAssetDecimals)
	var result []interface{}
	var path = []common.Address{tradePath.BaseAssetAddress, tradePath.QuoteAssetAddress}
	var err = u.routerContract.Call(&bind.CallOpts{}, &result, "getAmountsOut", amountInParsed, path)

	if err != nil {
		helpers.VerboseLog(verbose, fmt.Sprintf("Error getting price for %s: %v", symbol.Symbol, err))
		return 0
	}

	var priceList = result[0].([]*big.Int)
	return ethersHelper.WeiToEther(priceList[len(priceList)-1], tradePath.QuoteAssetDecimals)
}

func (u *PancakeswapWeb3Service) GetPriceMultiplePaths(
	tradePaths []sp.TradePath,
	amountIn float64,
	verbose bool,
) float64 {
	var path = []common.Address{tradePaths[0].BaseAssetAddress}
	for _, tradePath := range tradePaths {
		path = append(path, tradePath.QuoteAssetAddress)
	}
	var amountInParsed = ethersHelper.EtherToWei(amountIn, tradePaths[0].BaseAssetDecimals)
	var result []interface{}
	var err = u.routerContract.Call(&bind.CallOpts{}, &result, "getAmountsOut", amountInParsed, path)

	if err != nil {
		helpers.VerboseLog(verbose, fmt.Sprintf("Error getting price: %v", err))
		return 0
	}

	var priceList = result[0].([]*big.Int)
	return ethersHelper.WeiToEther(priceList[len(priceList)-1], tradePaths[len(tradePaths)-1].QuoteAssetDecimals)
}

// AggregatePrices ... aggregates prices
func (u *PancakeswapWeb3Service) AggregatePrices(symbols []*sp.Symbol, verbose bool) *sync.Map {
	var channel = make(chan *sp.Symbol)
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

// GetPoolDataByIndex ... get pool address by index (in factory) then get pool data
func (u *PancakeswapWeb3Service) GetPoolDataByIndex(index int) (sp.Symbol, error) {
	var result []interface{}
	var err = u.factoryContract.Call(&bind.CallOpts{}, &result, "allPairs", big.NewInt(int64(index)))
	helpers.Panic(err)
	var poolAddress = result[0].(common.Address)

	return u.GetPoolData(poolAddress)
}

func (u *PancakeswapWeb3Service) GetPoolData(address common.Address) (sp.Symbol, error) {
	poolABI, err := jsonHelper.ReadJSONABIFile("data/web3/pancakeswapPoolABI.json")
	helpers.Panic(err)
	erc20ABI, err := jsonHelper.ReadJSONABIFile("data/web3/erc20.json")
	helpers.Panic(err)

	var poolContract = bind.NewBoundContract(address, poolABI, u.client, u.client, u.client)
	var wg = sync.WaitGroup{}
	wg.Add(2)
	var symbol sp.Symbol
	var resultToken0, resultToken1 []interface{}
	var errToken0, errToken1 error
	go ethersHelper.CallContractMethod(
		&wg, poolContract, "token0", []interface{}{}, &resultToken0, &errToken0,
	)
	go ethersHelper.CallContractMethod(
		&wg, poolContract, "token1", []interface{}{}, &resultToken1, &errToken1,
	)
	//go ethers.CallContractMethod(&wg, poolContract, "fee", []interface{}{}, &resultFee, &errFee)
	wg.Wait()

	if errToken0 != nil || errToken1 != nil {
		return symbol, fmt.Errorf("error getting token addresses: %v, %v", errToken0, errToken1)
	}

	for _, result := range [][]interface{}{resultToken0, resultToken1} {
		wg.Add(2)
		var tokenAddress = result[0].(common.Address)
		var tokenContract = bind.NewBoundContract(tokenAddress, erc20ABI, u.client, u.client, u.client)
		var resultSymbol, resultDecimals []interface{}
		var errSymbol, errDecimals error
		go ethersHelper.CallContractMethod(
			&wg, tokenContract, "symbol", []interface{}{}, &resultSymbol, &errSymbol,
		)
		go ethersHelper.CallContractMethod(
			&wg, tokenContract, "decimals", []interface{}{}, &resultDecimals, &errDecimals,
		)
		wg.Wait()

		if errSymbol != nil || errDecimals != nil {
			return symbol, fmt.Errorf("error getting token addresses: %v, %v", errSymbol, errDecimals)
		}

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

	return symbol, nil
}
