//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package coinbase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/svanas/ladder/flag"
)

type Client struct {
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

func format(path string) string {
	if strings.HasPrefix(path, "/api") {
		return path
	} else {
		return fmt.Sprintf("/api/%s/brokerage/%s", apiVersion, path)
	}
}

func (self *Client) get(path string, values *url.Values) ([]byte, error) {
	beforeRequest()
	defer afterRequest()

	request, err := http.NewRequest("GET", func() string {
		result := apiBase + format(path)
		if values != nil {
			result += "?" + values.Encode()
		}
		return result
	}(), nil)
	if err != nil {
		return nil, err
	}

	nonce := strconv.FormatInt((time.Now().UTC().Unix()), 10)

	request.Header.Add("CB-ACCESS-KEY", self.apiKey)
	request.Header.Add("CB-ACCESS-TIMESTAMP", nonce)

	mac := hmac.New(sha256.New, []byte(self.apiSecret))
	if _, err := mac.Write([]byte(nonce + "GET" + format(path))); err != nil {
		return nil, err
	}
	request.Header.Add("CB-ACCESS-SIGN", hex.EncodeToString(mac.Sum(nil)))

	response, err := self.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		type Error struct {
			Message string `json:"message"`
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

func New() (*Client, error) {
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
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
