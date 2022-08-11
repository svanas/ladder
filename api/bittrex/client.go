//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package bittrex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	apiKey     string
	apiSecret  string
	appId      string
	httpClient *http.Client
	markets    []Market
}

func (self *Client) Do(method string, path string, payload []byte, auth bool) ([]byte, error) {
	var (
		code int
		out  []byte
		err  error
	)
	for {
		code, out, err = self.do(method, path, payload, auth)
		if code != http.StatusTooManyRequests {
			break
		}
	}
	return out, err
}

func (self *Client) do(method string, path string, payload []byte, auth bool) (int, []byte, error) {
	cooled, err := beforeRequest(path)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		afterRequest()
	}()

	url := func() string {
		if strings.HasPrefix(path, "http") {
			return path
		} else {
			return fmt.Sprintf("%s/%s/%s", apiBase, apiVersion, path)
		}
	}()

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:96.0) Gecko/20100101 Firefox/96.0")

	if self.appId != "" {
		req.Header.Add("Application-Id", self.appId)
	}

	if auth {
		// Unix timestamp in millisecond format
		nonce := strconv.FormatInt((time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)), 10)

		req.Header.Add("Api-Key", self.apiKey)
		req.Header.Add("Api-Timestamp", nonce)

		hash := sha512.New()
		if _, err := hash.Write([]byte(payload)); err != nil {
			return 0, nil, err
		}
		content := hex.EncodeToString(hash.Sum(nil))
		req.Header.Add("Api-Content-Hash", content)

		mac := hmac.New(sha512.New, []byte(self.apiSecret))
		if _, err := mac.Write([]byte(nonce + url + method + content)); err != nil {
			return 0, nil, err
		}
		req.Header.Add("Api-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := self.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusTooManyRequests {
			handleRateLimitErr(path, cooled)
		}
		return resp.StatusCode, nil, func() error {
			pair := make(map[string]string)
			json.Unmarshal(out, &pair)
			if msg, ok := pair["code"]; ok {
				return errors.New(msg)
			} else {
				return errors.New(resp.Status)
			}
		}()
	}

	return resp.StatusCode, out, nil
}

func (self *Client) getMarkets(cached bool) ([]Market, error) {
	if self.markets == nil || !cached {
		markets, err := func() (markets []Market, err error) {
			data, err := self.Do("GET", "markets", nil, false)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(data, &markets); err != nil {
				return nil, err
			}
			return markets, err
		}()
		if err != nil {
			return nil, err
		}
		self.markets = nil
		for _, market := range markets {
			if market.Active() {
				self.markets = append(self.markets, market)
			}
		}
	}
	return self.markets, nil
}

func (self *Client) GetMarket(name string) (*Market, error) {
	cached := true
	for {
		markets, err := self.getMarkets(cached)
		if err != nil {
			return nil, err
		}
		for _, market := range markets {
			if market.Symbol == name {
				return &market, nil
			}
		}
		if !cached {
			return nil, fmt.Errorf("market %s does not exist", name)
		}
		cached = false
	}
}

func (self *Client) CreateOrder(
	market string,
	direction consts.OrderSide,
	orderType consts.OrderType,
	quantity float64,
	limit float64,
	timeInForce consts.TimeInForce,
) (out *Order, err error) {
	type order struct {
		MarketSymbol string `json:"marketSymbol"`
		Direction    string `json:"direction"` // BUY or SELL
		OrderType    string `json:"type"`      // LIMIT or MARKET
		Quantity     string `json:"quantity"`
		Limit        string `json:"limit,omitempty"`
		TimeInForce  string `json:"timeInForce"`
	}

	new := &order{
		MarketSymbol: market,
		Direction:    direction.String(),
		OrderType:    orderType.String(),
		Quantity:     strconv.FormatFloat(quantity, 'f', -1, 64),
		TimeInForce:  timeInForce.String(),
	}

	if limit > 0 {
		new.Limit = strconv.FormatFloat(limit, 'f', -1, 64)
	}

	payload, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}

	data, err := self.Do("POST", "orders", payload, true)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (self *Client) CancelOrder(orderId string) error {
	_, err := self.Do("DELETE", fmt.Sprintf("orders/%s", orderId), nil, true)
	return err
}

func (self *Client) GetOpenOrders(market string) (orders []Order, err error) {
	data, err := self.Do("GET", fmt.Sprintf("orders/open?marketSymbol=%s", market), nil, true)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (self *Client) GetTicker(market string) (out *Ticker, err error) {
	data, err := self.Do("GET", fmt.Sprintf("markets/%s/ticker", market), nil, false)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

func ReadOnly() *Client {
	return &Client{
		"",
		"",
		appId,
		&http.Client{
			Timeout: 30 * time.Second,
		},
		nil,
	}
}

func ReadWrite() (*Client, error) {
	apiKey, err := flag.ApiKey()
	if err != nil {
		return nil, err
	}

	apiSecret, err := flag.ApiSecret()
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey,
		apiSecret,
		appId,
		&http.Client{
			Timeout: 30 * time.Second,
		},
		nil,
	}, nil
}
