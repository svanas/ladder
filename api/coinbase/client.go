//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package coinbase

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/svanas/ladder/flag"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type Client struct {
	apiKey     string
	apiSecret  string
	httpClient http.Client
}

type nonceSource struct{}

func (n nonceSource) Nonce() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

func format(path string) string {
	if strings.HasPrefix(path, "/api") {
		return path
	} else {
		return fmt.Sprintf("/api/%s/brokerage/%s", apiVersion, path)
	}
}

func (self *Client) auth(request *http.Request, method, path, body string) error {
	if strings.HasPrefix(self.apiKey, "organizations/") {
		// (coinbase developer platform) JWT key
		block, _ := pem.Decode([]byte(strings.Replace(self.apiSecret, "\\n", "\n", -1)))
		if block == nil {
			return errors.New("could not decode private key, you might need to enclose it in double quote characters")
		}

		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return err
		}

		sig, err := jose.NewSigner(
			jose.SigningKey{Algorithm: jose.ES256, Key: key},
			(&jose.SignerOptions{NonceSource: nonceSource{}}).WithType("JWT").WithHeader("kid", self.apiKey),
		)
		if err != nil {
			return err
		}

		type apiKeyClaims struct {
			*jwt.Claims
			URI string `json:"uri"`
		}

		claims := &apiKeyClaims{
			Claims: &jwt.Claims{
				Subject:   self.apiKey,
				Issuer:    "coinbase-cloud",
				NotBefore: jwt.NewNumericDate(time.Now()),
				Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
			},
			URI: fmt.Sprintf("%s %s%s", method, "api.coinbase.com", format(path)),
		}

		bearer, err := jwt.Signed(sig).Claims(claims).CompactSerialize()
		if err != nil {
			return err
		}
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearer))
	} else {
		// legacy API key
		nonce := strconv.FormatInt((time.Now().UTC().Unix()), 10)

		request.Header.Add("CB-ACCESS-KEY", self.apiKey)
		request.Header.Add("CB-ACCESS-TIMESTAMP", nonce)

		mac := hmac.New(sha256.New, []byte(self.apiSecret))
		if _, err := mac.Write([]byte(nonce + method + format(path) + body)); err != nil {
			return err
		}
		request.Header.Add("CB-ACCESS-SIGN", hex.EncodeToString(mac.Sum(nil)))
	}

	return nil
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

	if err := self.auth(request, "GET", path, ""); err != nil {
		return nil, err
	}

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

	if err := self.auth(request, "POST", path, func() string {
		if body != nil {
			return string(body)
		}
		return ""
	}()); err != nil {
		return nil, err
	}

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
