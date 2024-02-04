package coingecko

import (
	"fmt"
)

const (
	apiBase    = "https://api.coingecko.com/api/"
	apiVersion = "v3/"
)

func apiRequestsPerSecond() float64 {
	return float64(5) / float64(60) // 5 req/minute
}

func chainName(chainId int64) (string, error) {
	switch chainId {
	case 1:
		return "ethereum", nil
	case 10:
		return "optimistic-ethereum", nil
	case 56:
		return "binance-smart-chain", nil
	case 137:
		return "polygon-pos", nil
	case 250:
		return "fantom", nil
	case 8453:
		return "base", nil
	case 42161:
		return "arbitrum-one", nil
	case 43114:
		return "avalanche", nil
	}
	return "", fmt.Errorf("chain %d does not exist", chainId)
}
