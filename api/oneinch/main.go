package oneinch

import (
	"bytes"
	_ "embed"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/svanas/ladder/api/web3"
)

const (
	apiURL    = "https://api.1inch.com"
	apiRouter = "0x111111125421cA6dc452d289314280a0f8842A65"
)

//go:embed 1inch.api.key
var apiKey string

//go:embed router.abi.json
var apiRouterABI []byte

func getEpoch(chainId int64, maker common.Address) (*big.Int, error) {
	web3, err := web3.New(chainId)
	if err != nil {
		return nil, err
	}

	abi, err := abi.JSON(bytes.NewReader(apiRouterABI))
	if err != nil {
		return nil, err
	}

	data, err := abi.Pack("epoch", maker, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	to := common.HexToAddress(apiRouter)
	response, err := web3.Call(ethereum.CallMsg{To: &to, Data: data}, nil)
	if err != nil {
		return nil, err
	}

	var epoch *big.Int
	if err := abi.UnpackIntoInterface(&epoch, "epoch", response); err != nil {
		return nil, err
	}

	return epoch, nil
}
