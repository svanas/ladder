package paraswap

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	baseURL    string
	ChainId    int64
	privateKey []byte
	httpClient http.Client
}

func (client *Client) ecdsaPrivateKey() (*ecdsa.PrivateKey, error) {
	if client.privateKey == nil {
		return nil, fmt.Errorf("--%s cannot be empty", consts.FLAG_PRIVATE_KEY)
	}
	return crypto.ToECDSA(client.privateKey)
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

func (client *Client) do(request http.Request) ([]byte, error) {
	response, err := client.httpClient.Do(&request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		type Error struct {
			Message string `json:"error"`
		}
		var error Error
		if json.Unmarshal(body, &error) == nil {
			return nil, errors.New(error.Message)
		} else {
			return nil, errors.New(response.Status)
		}
	}

	return body, nil
}

func (client *Client) get(path string) ([]byte, error) {
	request, err := http.NewRequest("GET", (client.baseURL + path), nil)
	if err != nil {
		return nil, err
	}
	return client.do(*request)
}

func (client *Client) post(path string, body []byte) ([]byte, error) {
	request, err := http.NewRequest("POST", (client.baseURL + path), func() io.Reader {
		if body != nil {
			return bytes.NewReader(body)
		}
		return nil
	}())
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Add("Content-Type", "application/json")
	}
	return client.do(*request)
}

func ReadOnly() (*Client, error) {
	chainId, err := flag.ChainId()
	if err != nil {
		return nil, err
	}
	return &Client{
		apiBase,
		chainId,
		nil,
		http.Client{
			Timeout: 30 * time.Second,
		},
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
		apiBase,
		chainId,
		privateKey,
		http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
