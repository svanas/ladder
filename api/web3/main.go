package web3

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Client struct {
	client *geth.Client
}

//go:embed infura.api.key
var apiKey string

//go:embed erc20.abi.json
var erc20 []byte

func endpoint(chainId int64) (string, error) {
	switch chainId {
	case 1: // Ethereum
		return fmt.Sprintf("https://mainnet.infura.io/v3/%s", apiKey), nil
	case 10: // Optimism
		return fmt.Sprintf("https://optimism-mainnet.infura.io/v3/%s", apiKey), nil
	case 56: // Binance Smart Chain
		return "https://bsc-dataseed.binance.org", nil
	case 137: // Polygon
		return fmt.Sprintf("https://polygon-mainnet.infura.io/v3/%s", apiKey), nil
	case 250: // Fantom
		return "https://rpc.fantom.network", nil
	case 8453: // Base
		return "https://mainnet.base.org", nil
	case 42161: // Arbitrum
		return fmt.Sprintf("https://arbitrum-mainnet.infura.io/v3/%s", apiKey), nil
	case 43114: // Avalanche
		return fmt.Sprintf("https://avalanche-mainnet.infura.io/v3/%s", apiKey), nil
	}
	return "", fmt.Errorf("chain %d does not exist", chainId)
}

func New(chainId int64) (*Client, error) {
	url, err := endpoint(chainId)
	if err != nil {
		return nil, err
	}
	client, err := geth.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

func (client *Client) call(msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
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
	response, err := client.call(ethereum.CallMsg{
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
	response, err := client.call(ethereum.CallMsg{
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
