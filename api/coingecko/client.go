package coingecko

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type Details struct {
	Decimals int `json:"decimal_place"`
}

type Ticker struct {
	Last struct {
		Usd float64 `json:"usd"`
	} `json:"converted_last"`
}

type Coin struct {
	Id        string `json:"id"`
	Symbol    string `json:"symbol"`
	Platforms struct {
		Ethereum           string `json:"ethereum,omitempty"`
		OptimisticEthereum string `json:"optimistic-ethereum,omitempty"`
		BinanceSmartChain  string `json:"binance-smart-chain,omitempty"`
		PolygonPos         string `json:"polygon-pos,omitempty"`
		Fantom             string `json:"fantom,omitempty"`
		Base               string `json:"base,omitempty"`
		ArbitrumOne        string `json:"arbitrum-one,omitempty"`
		Avalanche          string `json:"avalanche,omitempty"`
	} `json:"platforms,omitempty"`
	Details struct {
		Ethereum           *Details `json:"ethereum,omitempty"`
		OptimisticEthereum *Details `json:"optimistic-ethereum,omitempty"`
		BinanceSmartChain  *Details `json:"binance-smart-chain,omitempty"`
		PolygonPos         *Details `json:"polygon-pos,omitempty"`
		Fantom             *Details `json:"fantom,omitempty"`
		Base               *Details `json:"base,omitempty"`
		ArbitrumOne        *Details `json:"arbitrum-one,omitempty"`
		Avalanche          *Details `json:"avalanche,omitempty"`
	} `json:"detail_platforms,omitempty"`
	Tickers []Ticker `json:"tickers,omitempty"`
}

type Client struct {
	baseURL    string
	httpClient http.Client
	coins      []Coin
	coin       map[string]Coin
}

func (client *Client) get(path string) ([]byte, error) {
	beforeRequest()
	defer afterRequest()

	response, err := client.httpClient.Get(client.baseURL + path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		type Error1 struct {
			Message string `json:"error"`
		}
		var err1 Error1
		if json.Unmarshal(body, &err1) == nil && err1.Message != "" {
			return nil, errors.New(err1.Message)
		}
		type Error2 struct {
			Status struct {
				Message string `json:"error_message"`
			} `json:"status"`
		}
		var err2 Error2
		if json.Unmarshal(body, &err2) == nil && err2.Status.Message != "" {
			return nil, errors.New(err2.Status.Message)
		}
		return nil, errors.New(response.Status)
	}

	return body, nil
}

func (client *Client) getCoins() ([]Coin, error) {
	if len(client.coins) == 0 {
		body, err := client.get("coins/list?include_platform=true")
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &client.coins); err != nil {
			return nil, err
		}
	}
	return client.coins, nil
}

func (client *Client) GetCoin(symbol string, chainId int64) (string, string, string, error) { // --> (coinId, symbol, address, error)
	chainName, err := chainName(chainId)
	if err != nil {
		return "", "", "", err
	}
	coins, err := client.getCoins()
	if err != nil {
		return "", "", "", err
	}
	for _, coin := range coins {
		if strings.EqualFold(coin.Symbol, symbol) || (len(symbol) == 42 && strings.HasPrefix(strings.ToLower(symbol), "0x")) {
			v := reflect.ValueOf(coin.Platforms)
			for i := 0; i < v.NumField(); i++ {
				if strings.EqualFold(v.Type().Field(i).Name, strings.ReplaceAll(chainName, "-", "")) {
					address := v.Field(i).String()
					if address != "" && (strings.EqualFold(coin.Symbol, symbol) || strings.EqualFold(symbol, address)) {
						return coin.Id, coin.Symbol, address, nil
					}
				}
			}
		}
	}
	return "", "", "", fmt.Errorf("token %s does not exist on chain %d", symbol, chainId)
}

func (client *Client) getCoin(coinId string) (*Coin, error) {
	coin, ok := client.coin[coinId]
	if ok {
		return &coin, nil
	}
	body, err := client.get("coins/" + coinId)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &coin); err != nil {
		return nil, err
	}
	client.coin[coinId] = coin
	return &coin, nil
}

func (client *Client) GetDecimals(coinId string, chainId int64) (int, error) {
	coin, err := client.getCoin(coinId)
	if err != nil {
		return 0, err
	}
	chainName, err := chainName(chainId)
	if err != nil {
		return 0, err
	}
	v := reflect.ValueOf(coin.Details)
	for i := 0; i < v.NumField(); i++ {
		if strings.EqualFold(v.Type().Field(i).Name, strings.ReplaceAll(chainName, "-", "")) {
			details, ok := v.Field(i).Interface().(*Details)
			if !ok {
				return 0, fmt.Errorf("error casting %s.%s's interface to details", coinId, chainName)
			}
			return details.Decimals, nil
		}
	}
	return 0, fmt.Errorf("%s's decimals not found on chain %s", coinId, chainName)
}

func (client *Client) GetTicker(coinId string) (float64, error) {
	coin, err := client.getCoin(coinId)
	if err != nil {
		return 0, err
	}
	if len(coin.Tickers) == 0 {
		return 0, fmt.Errorf("%s's ticker not found", coinId)
	}
	return coin.Tickers[0].Last.Usd, nil
}

func New() *Client {
	return &Client{
		apiBase + apiVersion,
		http.Client{
			Timeout: 30 * time.Second,
		},
		nil,
		map[string]Coin{},
	}
}
