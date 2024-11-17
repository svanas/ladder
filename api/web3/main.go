package web3

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client *geth.Client
}

//go:embed infura.api.key
var apiKey string

//go:embed erc20.abi.json
var erc20 []byte

const (
	Ethereum          int64 = 1
	Optimism          int64 = 10
	BinanceSmartChain int64 = 56
	Polygon           int64 = 137
	Fantom            int64 = 250
	Base              int64 = 8453
	Arbitrum          int64 = 42161
	Avalanche         int64 = 43114
)

var Chains = [8]int64{Ethereum, Optimism, BinanceSmartChain, Polygon, Fantom, Base, Arbitrum, Avalanche}

func Endpoint(chainId int64) (string, error) {
	if apiKey == "" {
		return "", errors.New("please generate yourself an API key on infura.io then paste your API key in infura.api.key and recompile")
	}

	switch chainId {
	case Ethereum:
		return fmt.Sprintf("https://mainnet.infura.io/v3/%s", apiKey), nil
	case Optimism:
		return fmt.Sprintf("https://optimism-mainnet.infura.io/v3/%s", apiKey), nil
	case BinanceSmartChain:
		return "https://bsc-dataseed.binance.org", nil
	case Polygon:
		return fmt.Sprintf("https://polygon-mainnet.infura.io/v3/%s", apiKey), nil
	case Fantom:
		return "https://rpc.fantom.network", nil
	case Base:
		return "https://mainnet.base.org", nil
	case Arbitrum:
		return fmt.Sprintf("https://arbitrum-mainnet.infura.io/v3/%s", apiKey), nil
	case Avalanche:
		return fmt.Sprintf("https://avalanche-mainnet.infura.io/v3/%s", apiKey), nil
	}

	return "", fmt.Errorf("chain %d is not supported at this time", chainId)
}

func Checksum(address string) string {
	return common.HexToAddress(address).Hex()
}

func New(chainId int64) (*Client, error) {
	url, err := Endpoint(chainId)
	if err != nil {
		return nil, err
	}
	client, err := geth.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (client *Client) Call(msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return client.client.CallContract(context.Background(), msg, blockNumber)
}

func (client *Client) GetSymbol(contract string) (string, error) {
	return client.getSymbol(common.HexToAddress(contract))
}

func (client *Client) getSymbol(contract common.Address) (string, error) {
	parsed, err := abi.JSON(bytes.NewReader(erc20))
	if err != nil {
		return "", err
	}

	// query the chain
	response, err := client.Call(ethereum.CallMsg{
		To:   &contract,
		Data: parsed.Methods["symbol"].ID,
	}, nil)
	if err != nil {
		return "", err
	}

	// unpack the result
	var symbol string
	if err := parsed.UnpackIntoInterface(&symbol, "symbol", response); err != nil {
		return "", err
	}

	return symbol, nil
}

func (client *Client) GetAllowance(contract, owner, spender string) (*big.Int, error) {
	return client.getAllowance(
		common.HexToAddress(contract),
		common.HexToAddress(owner),
		common.HexToAddress(spender),
	)
}

func (client *Client) getAllowance(contract, owner, spender common.Address) (*big.Int, error) {
	parsed, err := abi.JSON(bytes.NewReader(erc20))
	if err != nil {
		return nil, err
	}

	data, err := parsed.Pack("allowance", owner, spender)
	if err != nil {
		return nil, err
	}

	// query the chain
	response, err := client.Call(ethereum.CallMsg{
		To:   &contract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	// unpack the result
	var allowance *big.Int
	if err := parsed.UnpackIntoInterface(&allowance, "allowance", response); err != nil {
		return nil, err
	}

	return allowance, nil
}

func (client *Client) GetDecimals(contract string) (int, error) {
	return client.getDecimals(common.HexToAddress(contract))
}

func (client *Client) getDecimals(contract common.Address) (int, error) {
	parsed, err := abi.JSON(bytes.NewReader(erc20))
	if err != nil {
		return 0, err
	}

	// query the chain
	response, err := client.Call(ethereum.CallMsg{
		To:   &contract,
		Data: parsed.Methods["decimals"].ID,
	}, nil)
	if err != nil {
		return 0, err
	}

	// unpack the result
	var decimals uint8
	if err := parsed.UnpackIntoInterface(&decimals, "decimals", response); err != nil {
		return 0, err
	}

	return int(decimals), nil
}
