package ethers

import (
	"math"
	"math/big"
)

func WeiToEther(wei *big.Int, decimals int) float64 {
	f := new(big.Float).SetInt(wei)
	ethValue, _ := f.Quo(f, big.NewFloat(math.Pow10(decimals))).Float64()
	return ethValue
}
