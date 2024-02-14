//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package coinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	httpClient http.Client
}

func format(path string) string {
	if strings.HasPrefix(path, "/api") {
		return path
	} else {
		return fmt.Sprintf("/api/%s/brokerage/%s", apiVersion, path)
	}
}

func (self *Client) do(request http.Request) ([]byte, error) {
	response, err := self.httpClient.Do(&request)
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

	return self.do(*request)
}

func (self *Client) post(path string, body []byte) ([]byte, error) {
	beforeRequest()
	defer afterRequest()

	request, err := http.NewRequest("POST", (apiBase + format(path)), func() io.Reader {
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

	nonce := strconv.FormatInt((time.Now().UTC().Unix()), 10)

	request.Header.Add("CB-ACCESS-KEY", self.apiKey)
	request.Header.Add("CB-ACCESS-TIMESTAMP", nonce)

	mac := hmac.New(sha256.New, []byte(self.apiSecret))
	if _, err := mac.Write([]byte(nonce + "POST" + format(path) + func() string {
		if body != nil {
			return string(body)
		}
		return ""
	}())); err != nil {
		return nil, err
	}
	request.Header.Add("CB-ACCESS-SIGN", hex.EncodeToString(mac.Sum(nil)))

	return self.do(*request)
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
		http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
