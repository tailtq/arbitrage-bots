package web3

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"sync"
)

func GetPancakeSwapAddresses(networkName string) map[string]string {
	if networkName == "bsc" {
		return map[string]string{
			"factory": "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73",
			"router":  "0x10ED43C718714eb63d5aA57B78B54704E256024E",
		}
	}
	return map[string]string{}
}

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
