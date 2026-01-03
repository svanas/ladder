package oneinch

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	ChainId    int64
	privateKey []byte
	httpClient http.Client
}

func (client *Client) do(request http.Request) ([]byte, error) {
	beforeRequest()
	defer afterRequest()

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

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
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		var error Error
		if json.Unmarshal(body, &error) == nil {
			return nil, errors.New(func() string {
				msg := strings.TrimSpace(error.Error)
				if error.Message != "" {
					if msg != "" {
						if !strings.HasSuffix(msg, ".") {
							msg += ". "
						} else {
							msg += " "
						}
					}
					msg += error.Message
				}
				return msg
			}())
		} else {
			return nil, errors.New(response.Status)
		}
	}

	return body, nil
}

func (client *Client) get(path string) ([]byte, error) {
	request, err := http.NewRequest("GET", (apiURL + path), nil)
	if err != nil {
		return nil, err
	}
	return client.do(*request)
}

func (client *Client) post(path string, body []byte) ([]byte, error) {
	request, err := http.NewRequest("POST", (apiURL + path), func() io.Reader {
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

func (client *Client) ecdsaPrivateKey() (*ecdsa.PrivateKey, error) {
	if client.privateKey == nil {
		return nil, fmt.Errorf("--%s cannot be empty", consts.FLAG_PRIVATE_KEY)
	}
	return crypto.ToECDSA(client.privateKey)
}

func (client *Client) publicAddress() (common.Address, error) {
	if client.privateKey == nil {
		return [20]byte{}, fmt.Errorf("--%s cannot be empty", consts.FLAG_PRIVATE_KEY)
	}
	ecdsaPrivateKey, err := crypto.ToECDSA(client.privateKey)
	if err != nil {
		return [20]byte{}, err
	}
	return crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey), nil
}

func (client *Client) GetEpoch() (*big.Int, error) {
	maker, err := client.publicAddress()
	if err != nil {
		return big.NewInt(0), err
	}
	return getEpoch(client.ChainId, maker)
}

func ReadOnly() (*Client, error) {
	chainId, err := flag.ChainId()
	if err != nil {
		return nil, err
	}
	return &Client{
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
		chainId,
		privateKey,
		http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
