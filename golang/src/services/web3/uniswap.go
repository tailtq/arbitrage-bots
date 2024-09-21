package web3

import (
	"arbitrage-bot/helpers"
	"arbitrage-bot/helpers/ethers"
	ethersHelper "arbitrage-bot/helpers/ethers"
	jsonHelper "arbitrage-bot/helpers/json"
	"arbitrage-bot/services/sourceprovider"
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
	"sync"
)

type UniswapWeb3Service struct {
	client         *ethclient.Client
	quoterAddress  common.Address
	quoterABI      abi.ABI
	quoterVersion  string
	quoterContract *bind.BoundContract
}

func NewUniswapWeb3Service() *UniswapWeb3Service {
	var networkRpcUrl = os.Getenv("NETWORK_RPC_URL")
	var networkName = os.Getenv("NETWORK_NAME")
	var quoterAddress = common.HexToAddress(os.Getenv("UNISWAP_QUOTER_ADDRESS"))
	var quoterABI abi.ABI
	var quoterVersion string
	var err error

	if networkName == "blast" || networkName == "celo" {
		// use quoterV2
		quoterABI, err = jsonHelper.ReadJSONABIFile("data/web3/uniswapQuoterV2ABI.json")
		quoterVersion = "v2"
	} else {
		quoterABI, err = jsonHelper.ReadJSONABIFile("data/web3/uniswapQuoterABI.json")
		quoterVersion = "v1"
	}
	helpers.Panic(err)
	client, err := ethclient.Dial(networkRpcUrl)
	helpers.Panic(err)

	return &UniswapWeb3Service{
		client:        client,
		quoterAddress: quoterAddress,
		quoterABI:     quoterABI,
		quoterVersion: quoterVersion,
		// not used yet
		//quoterContract: bind.NewBoundContract(quoterAddress, quoterABI, client, client, client),
	}
}

// GetPoolData ... returns the pool data for a given pool address
func (u *UniswapWeb3Service) GetPoolData(address common.Address) sourceprovider.Symbol {
	poolABI, err := jsonHelper.ReadJSONABIFile("data/web3/uniswapPoolABI.json")
	helpers.Panic(err)
	erc20ABI, err := jsonHelper.ReadJSONABIFile("data/web3/erc20.json")
	helpers.Panic(err)

	var poolContract = bind.NewBoundContract(address, poolABI, u.client, u.client, u.client)
	var wg = sync.WaitGroup{}
	wg.Add(3)
	var symbol sourceprovider.Symbol
	var resultToken0, resultToken1, resultFee []interface{}
	var errToken0, errToken1, errFee error
	go ethers.CallContractMethod(&wg, poolContract, "token0", []interface{}{}, &resultToken0, &errToken0)
	go ethers.CallContractMethod(&wg, poolContract, "token1", []interface{}{}, &resultToken1, &errToken1)
	go ethers.CallContractMethod(&wg, poolContract, "fee", []interface{}{}, &resultFee, &errFee)
	wg.Wait()
	helpers.PanicBatch(errToken0, errToken1, errFee)

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
	symbol.FeeTier = int(resultFee[0].(*big.Int).Int64())

	return symbol
}

// GetPrice ... returns the price for a given symbol
func (u *UniswapWeb3Service) GetPrice(symbol sourceprovider.Symbol, amountIn float64, tradeDirection string, verbose bool) float64 {
	var tradePath = ethersHelper.GetTradePaths([]sourceprovider.Symbol{symbol}, []string{tradeDirection})[0]

	if u.quoterVersion == "v2" {
		return u.quoteExactInputSingleV2(
			symbol,
			amountIn,
			tradePath.BaseAssetAddress,
			tradePath.QuoteAssetAddress,
			tradePath.BaseAssetDecimals,
			tradePath.QuoteAssetDecimals,
			verbose,
		)
	} else {
		return u.quoteExactInputSingleV1(
			symbol,
			amountIn,
			tradePath.BaseAssetAddress,
			tradePath.QuoteAssetAddress,
			tradePath.BaseAssetDecimals,
			tradePath.QuoteAssetDecimals,
			verbose,
		)
	}
}

func (u *UniswapWeb3Service) GetPriceMultiplePaths(
	symbols []sourceprovider.Symbol,
	tradeDirections []string,
	amountIn float64,
	verbose bool,
) float64 {
	return 0
}

func (u *UniswapWeb3Service) quoteExactInputSingleV1(
	symbol sourceprovider.Symbol,
	amountIn float64,
	inputTokenA common.Address,
	inputTokenB common.Address,
	inputDecimalsA int,
	inputDecimalsB int,
	verbose bool,
) float64 {
	var amountInParsed = big.NewInt(int64(amountIn * math.Pow(10, float64(inputDecimalsA))))
	data, err := u.quoterABI.Pack(
		"quoteExactInputSingle",
		inputTokenA,
		inputTokenB,
		big.NewInt(int64(symbol.FeeTier)),
		amountInParsed,
		big.NewInt(0),
	)
	helpers.Panic(err)

	var message = ethereum.CallMsg{To: &u.quoterAddress, Data: data}
	result, err := u.client.CallContract(context.Background(), message, nil)

	if err != nil {
		if verbose {
			fmt.Println("Quoter error:", symbol.Symbol, err)
		}
		return 0
	}
	var quotedAmountOut = new(big.Int)
	quotedAmountOut.SetBytes(result)
	return ethers.WeiToEther(quotedAmountOut, inputDecimalsB)
}

func (u *UniswapWeb3Service) quoteExactInputSingleV2(
	symbol sourceprovider.Symbol,
	amountIn float64,
	inputTokenA common.Address,
	inputTokenB common.Address,
	inputDecimalsA int,
	inputDecimalsB int,
	verbose bool,
) float64 {
	type QuoteExactInputSingleParams struct {
		TokenIn           common.Address `json:"tokenIn"`
		TokenOut          common.Address `json:"tokenOut"`
		AmountIn          *big.Int       `json:"amountIn"`
		Fee               *big.Int       `json:"fee"`
		SqrtPriceLimitX96 *big.Int       `json:"sqrtPriceLimitX96"`
	}

	var amountInParsed = big.NewInt(int64(amountIn * math.Pow(10, float64(inputDecimalsA))))
	data, err := u.quoterABI.Pack(
		"quoteExactInputSingle",
		QuoteExactInputSingleParams{
			inputTokenA,
			inputTokenB,
			amountInParsed,
			big.NewInt(int64(symbol.FeeTier)),
			big.NewInt(0),
		},
	)
	helpers.Panic(err)

	var message = ethereum.CallMsg{To: &u.quoterAddress, Data: data}
	result, err := u.client.CallContract(context.Background(), message, nil)

	if err != nil {
		if verbose {
			fmt.Println("Quoter error:", symbol.Symbol, err)
		}
		return 0
	}

	// Unpack the result
	var amountOut *big.Int
	var sqrtPriceX96After *big.Int
	var initializedTicksCrossed uint32
	var gasEstimate *big.Int
	err = u.quoterABI.UnpackIntoInterface(&[]interface{}{
		&amountOut,
		&sqrtPriceX96After,
		&initializedTicksCrossed,
		&gasEstimate,
	}, "quoteExactInputSingle", result)

	return ethers.WeiToEther(amountOut, inputDecimalsB)
}

func (u *UniswapWeb3Service) AggregatePrices(symbols []*sourceprovider.Symbol, verbose bool) *sync.Map {
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
