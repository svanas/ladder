//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package bitstamp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/svanas/ladder/flag"
)

type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

var (
	cache []Pair
)

func (self *Client) reason(body []byte) error {
	resp := make(map[string]interface{})
	if json.Unmarshal(body, &resp) == nil {
		if reason1, ok := resp["reason"]; ok {
			if reason2, ok := reason1.(map[string]interface{}); ok {
				if all, ok := reason2["__all__"]; ok {
					msg := fmt.Sprintf("%v", all)
					if msg != "" && msg != "[]" {
						return errors.New(msg)
					}
				}
			}
			return fmt.Errorf("%v", reason1)
		}
	}
	return nil
}

func (self *Client) get(path string) ([]byte, error) {
	// satisfy the rate limiter (limited to 8000 requests per 10 minutes)
	beforeRequest("GET", path)
	defer afterRequest()

	// parse the bitstamp URL
	endpoint, err := url.Parse(self.baseURL)
	if err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path

	resp, err := self.httpClient.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s %s", resp.Status, path)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// is this an error?
	if err := self.reason(body); err != nil {
		return nil, err
	}

	return body, nil
}

func (self *Client) post(path string, values url.Values) ([]byte, error) {
	// satisfy the rate limiter (limited to 8000 requests per 10 minutes)
	beforeRequest("POST", path)
	defer afterRequest()

	// parse the bitstamp URL
	endpoint, err := url.Parse(self.baseURL)
	if err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path

	// encode the url.Values in the body
	payload := values.Encode()
	input := strings.NewReader(payload)

	// create the request
	req, err := http.NewRequest("POST", endpoint.String(), input)
	if err != nil {
		return nil, err
	}

	// there is no need to set Content-Type if there is no body
	if payload != "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// compute v2 authentication headers
	x_auth := "BITSTAMP" + " " + self.apiKey
	x_auth_nonce := func(length int) string {
		result := ""
		for len(result) < length {
			n, _ := rand.Int(rand.Reader, big.NewInt(10))
			result += n.String()
		}
		return strings.ToLower(result)
	}(36)
	x_auth_timestamp := strconv.FormatInt((time.Now().UnixNano() / 1000000), 10)
	x_auth_version := "v2"

	// v2 auth message that we will need to sign
	x_auth_message := x_auth +
		req.Method +
		req.Host +
		"/api/v2" + path +
		"" +
		func() string { // content_type
			if payload == "" {
				return ""
			}
			return "application/x-www-form-urlencoded"
		}() +
		x_auth_nonce +
		x_auth_timestamp +
		x_auth_version +
		payload

	// compute the v2 signature
	mac := hmac.New(sha256.New, []byte(self.apiSecret))
	mac.Write([]byte(x_auth_message))
	x_auth_signature := strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))

	// add v2 autentication headers
	req.Header.Add("X-Auth", x_auth)
	req.Header.Add("X-Auth-Nonce", x_auth_nonce)
	req.Header.Add("X-Auth-Timestamp", x_auth_timestamp)
	req.Header.Add("X-Auth-Version", x_auth_version)
	req.Header.Add("X-Auth-Signature", x_auth_signature)

	// submit the http request
	resp, err := self.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST %s %s", resp.Status, path)
	}

	// read the body of the http message into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// is this an error?
	if err := self.reason(body); err != nil {
		return nil, err
	}

	return body, nil
}

func (self *Client) getPairs(cached bool) ([]Pair, error) {
	if cache == nil || !cached {
		pairs, err := self.TradingPairsInfo()
		if err != nil {
			return nil, err
		}
		cache = nil
		for _, pair := range pairs {
			if strings.EqualFold(pair.Trading, "enabled") {
				cache = append(cache, pair)
			}
		}
	}
	return cache, nil
}

func (self *Client) GetPair(market string) (*Pair, error) {
	cached := true
	for {
		pairs, err := self.getPairs(cached)
		if err != nil {
			return nil, err
		}
		for _, pair := range pairs {
			if pair.UrlSymbol == market {
				return &pair, nil
			}
		}
		if !cached {
			return nil, fmt.Errorf("market %s does not exist", market)
		}
		cached = false
	}
}

func (self *Client) TradingPairsInfo() ([]Pair, error) {
	body, err := self.get("/trading-pairs-info/")
	if err != nil {
		return nil, err
	}

	var out []Pair
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (self *Client) GetOpenOrders(pair string) ([]Order, error) {
	body, err := self.post(fmt.Sprintf("/open_orders/%s/", pair), url.Values{})
	if err != nil {
		return nil, err
	}

	var out []Order
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (self *Client) CancelOrder(id string) error {
	values := url.Values{}
	values.Add("id", id)

	if _, err := self.post("/cancel_order/", values); err != nil {
		return err
	}

	return nil
}

func (self *Client) Ticker(pair string) (*Ticker, error) {
	body, err := self.get(fmt.Sprintf("/ticker/%s/", pair))
	if err != nil {
		return nil, err
	}
	var out Ticker
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (self *Client) BuyLimitOrder(pair string, amount, price float64) (*Order, error) {
	values := url.Values{}
	values.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	values.Add("price", strconv.FormatFloat(price, 'f', -1, 64))

	body, err := self.post(fmt.Sprintf("/buy/%s/", pair), values)
	if err != nil {
		return nil, err
	}

	var out Order
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (client *Client) SellLimitOrder(pair string, amount, price float64) (*Order, error) {
	values := url.Values{}
	values.Add("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	values.Add("price", strconv.FormatFloat(price, 'f', -1, 64))

	body, err := client.post(fmt.Sprintf("/sell/%s/", pair), values)
	if err != nil {
		return nil, err
	}

	var out Order
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func ReadOnly() *Client {
	return &Client{
		endpoint,
		"",
		"",
		&http.Client{
			Timeout: 30 * time.Second,
		},
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
		endpoint,
		apiKey,
		apiSecret,
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
