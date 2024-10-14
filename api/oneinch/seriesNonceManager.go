package oneinch

import (
	"bytes"
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/svanas/ladder/api/web3"
)

//go:embed seriesNonceManager.abi.json
var seriesNonceManagerABI []byte

func getSeriesNonceManager(chainId int64) (common.Address, error) {
	switch chainId {
	case web3.Arbitrum:
		return common.HexToAddress("0xD7936052D1e096d48C81Ef3918F9Fd6384108480"), nil
	case web3.Avalanche:
		return common.HexToAddress("0x2EC255797FEF7669fA243509b7a599121148FFba"), nil
	case web3.Base:
		return common.HexToAddress("0xD9Cc0A957cAC93135596f98c20Fbaca8Bf515909"), nil
	case web3.BinanceSmartChain:
		return common.HexToAddress("0x58ce0e6ef670c9a05622f4188faa03a9e12ee2e4"), nil
	case web3.Ethereum:
		return common.HexToAddress("0x303389f541ff2d620e42832f180a08e767b28e10"), nil
	case web3.Fantom:
		return common.HexToAddress("0x7871769b3816b23dB12E83a482aAc35F1FD35D4B"), nil
	case web3.Optimism:
		return common.HexToAddress("0x32d12a25f539E341089050E2d26794F041fC9dF8"), nil
	case web3.Polygon:
		return common.HexToAddress("0xa5eb255EF45dFb48B5d133d08833DEF69871691D"), nil
	default:
		return common.HexToAddress("0x0000000000000000000000000000000000000000"), fmt.Errorf("chain %d is not supported at this time", chainId)
	}
}

func getSeriesNonce(chainId int64, public common.Address) (*big.Int, error) {
	manager, err := getSeriesNonceManager(chainId)
	if err != nil {
		return nil, err
	}

	web3, err := web3.New(chainId)
	if err != nil {
		return nil, err
	}

	abi, err := abi.JSON(bytes.NewReader(seriesNonceManagerABI))
	if err != nil {
		return nil, err
	}

	data, err := abi.Pack("nonce", big.NewInt(0), public)
	if err != nil {
		return nil, err
	}

	response, err := web3.Call(ethereum.CallMsg{To: &manager, Data: data}, nil)
	if err != nil {
		return nil, err
	}

	var nonce *big.Int
	if err := abi.UnpackIntoInterface(&nonce, "nonce", response); err != nil {
		return nil, err
	}

	return nonce, nil
}
