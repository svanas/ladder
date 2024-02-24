package paraswap

import (
	"fmt"

	"github.com/svanas/ladder/api/web3"
)

const (
	apiBase = "https://api.paraswap.io/"
)

func router(chainId int64) (string, error) {
	switch chainId {
	case web3.Ethereum:
		return "0xe92b586627cca7a83dc919cc7127196d70f55a06", nil
	case web3.Optimism:
		return "0x0927fd43a7a87e3e8b81df2c44b03c4756849f6d", nil
	case web3.BinanceSmartChain:
		return "0x8dcdfe88ef0351f27437284d0710cd65b20288bb", nil
	case web3.Polygon:
		return "0xF3CD476C3C4D3Ac5cA2724767f269070CA09A043", nil
	case web3.Fantom:
		return "0x2df17455b96dde3618fd6b1c3a9aa06d6ab89347", nil
	case web3.Base:
		return "0xa003dFBA51C9e1e56C67ae445b852bdEd7aC5EEd", nil
	case web3.Arbitrum:
		return "0x0927fd43a7a87e3e8b81df2c44b03c4756849f6d", nil
	case web3.Avalanche:
		return "0x34302c4267d0da0a8c65510282cc22e9e39df51f", nil
	}
	return "", fmt.Errorf("chain %d is not supported at this time", chainId)
}
