package web3

import (
	"arbitrage-bot/helpers"
	"arbitrage-bot/helpers/ethers"
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

func callContractMethod(
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

type UniswapWeb3Service struct {
	client         *ethclient.Client
	quoterAddress  common.Address
	quoterABI      abi.ABI
	quoterContract *bind.BoundContract
}

func NewUniswapWeb3Service() *UniswapWeb3Service {
	var networkRpcUrl = os.Getenv("NETWORK_RPC_URL")
	var quoterAddress = common.HexToAddress(os.Getenv("UNISWAP_QUOTER_ADDRESS"))
	var quoterABI, err = jsonHelper.ReadJSONABIFile("data/web3/quoterABI.json")
	helpers.Panic(err)
	client, err := ethclient.Dial(networkRpcUrl)
	helpers.Panic(err)

	return &UniswapWeb3Service{
		client:         client,
		quoterAddress:  quoterAddress,
		quoterABI:      quoterABI,
		quoterContract: bind.NewBoundContract(quoterAddress, quoterABI, client, client, client),
	}
}

func (u *UniswapWeb3Service) GetPrice(symbol sourceprovider.Symbol, amountIn float64, tradeDirection string, verbose bool) float64 {
	if amountIn == 0 {
		return 0
	}

	var inputSymbolA, inputSymbolB string
	var inputTokenA, inputTokenB common.Address
	var inputDecimalsA, inputDecimalsB int
	_ = inputSymbolA
	_ = inputSymbolB
	_ = inputDecimalsB

	if tradeDirection == "baseToQuote" {
		inputTokenA = common.HexToAddress(symbol.BaseAssetAddress)
		inputSymbolA = symbol.BaseAsset
		inputDecimalsA = symbol.BaseAssetDecimals
		inputTokenB = common.HexToAddress(symbol.QuoteAssetAddress)
		inputSymbolB = symbol.QuoteAsset
		inputDecimalsB = symbol.QuoteAssetDecimals
	} else if tradeDirection == "quoteToBase" {
		inputTokenA = common.HexToAddress(symbol.QuoteAssetAddress)
		inputSymbolA = symbol.QuoteAsset
		inputDecimalsA = symbol.QuoteAssetDecimals
		inputTokenB = common.HexToAddress(symbol.BaseAssetAddress)
		inputSymbolB = symbol.BaseAsset
		inputDecimalsB = symbol.BaseAssetDecimals
	}

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

func (u *UniswapWeb3Service) AggregatePrices(symbols []*sourceprovider.Symbol, verbose bool) *sync.Map {
	var channel = make(chan *sourceprovider.Symbol)
	var result sync.Map
	var wg sync.WaitGroup

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for symbol := range channel {
				var price = u.GetPrice(*symbol, 1, "baseToQuote", verbose)
				result.Store(symbol.Symbol, price)
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
