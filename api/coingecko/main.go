package coingecko

import (
	"fmt"
	"github.com/svanas/ladder/api/web3"
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
	case web3.Ethereum:
		return "ethereum", nil
	case web3.Optimism:
		return "optimistic-ethereum", nil
	case web3.BinanceSmartChain:
		return "binance-smart-chain", nil
	case web3.Polygon:
		return "polygon-pos", nil
	case web3.Fantom:
		return "fantom", nil
	case web3.Base:
		return "base", nil
	case web3.Arbitrum:
		return "arbitrum-one", nil
	case web3.Avalanche:
		return "avalanche", nil
	}
	return "", fmt.Errorf("chain %d does not exist", chainId)
}
