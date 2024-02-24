package oneinch

import (
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	oneinch "github.com/svanas/1inch-sdk/golang/client"
	"github.com/svanas/ladder/api/web3"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	ChainId    int64
	privateKey []byte
}

//go:embed 1inch.api.key
var apiKey string

func (client *Client) oneInchConfig() (*oneinch.Config, error) {
	if apiKey == "" {
		return nil, errors.New("please generate yourself an API key on portal.1inch.dev then paste your API key in 1inch.api.key and recompile")
	}

	config := oneinch.Config{DevPortalApiKey: apiKey}

	for _, chainId := range web3.Chains {
		endpoint, err := web3.Endpoint(chainId)
		if err != nil {
			return nil, err
		}
		config.Web3HttpProviders = append(config.Web3HttpProviders, oneinch.Web3ProviderConfig{ChainId: int(chainId), Url: endpoint})
	}

	return &config, nil
}

func (client *Client) oneInchClient() (*oneinch.Client, error) {
	config, err := client.oneInchConfig()
	if err != nil {
		return nil, err
	}
	return oneinch.NewClient(*config)
}

func (client *Client) PrivateKey() (string, error) {
	if client.privateKey == nil {
		return "", fmt.Errorf("--%s cannot be empty", consts.FLAG_PRIVATE_KEY)
	}
	return hex.EncodeToString(client.privateKey), nil
}

func (client *Client) PublicAddress() (string, error) {
	if client.privateKey == nil {
		return "", fmt.Errorf("--%s cannot be empty", consts.FLAG_PRIVATE_KEY)
	}
	ecdsaPrivateKey, err := crypto.ToECDSA(client.privateKey)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey).Hex(), nil
}

func ReadOnly() (*Client, error) {
	chainId, err := flag.ChainId()
	if err != nil {
		return nil, err
	}
	return &Client{
		chainId,
		nil,
	}, nil
}

func ReadWrite() (*Client, error) {
	chainId, err := flag.ChainId()
	if err != nil {
		return nil, err
	}

	privateKey, err := flag.PrivateKey()
	if err != nil {
		return nil, err
	}

	return &Client{
		chainId,
		privateKey,
	}, nil
}
