//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package gdax

import (
	"fmt"
	"net/http"
	"time"

	coinbasepro "github.com/svanas/go-coinbasepro"
	"github.com/svanas/ladder/flag"
)

const (
	baseURL = "https://api.pro.coinbase.com"
	sandbox = "https://api-public.sandbox.pro.coinbase.com"
)

type Client struct {
	*coinbasepro.Client
	products []coinbasepro.Product
}

func (self *Client) getProducts(cached bool) ([]coinbasepro.Product, error) {
	if self.products == nil || !cached {
		products, err := self.Client.GetProducts()
		if err != nil {
			return nil, err
		}
		self.products = nil
		for _, product := range products {
			if !product.CancelOnly && !product.TradingDisabled {
				self.products = append(self.products, product)
			}
		}
	}
	return self.products, nil
}

func (self *Client) GetProduct(market string) (*coinbasepro.Product, error) {
	cached := true
	for {
		products, err := self.getProducts(cached)
		if err != nil {
			return nil, err
		}
		for _, product := range products {
			if product.ID == market {
				return &product, nil
			}
		}
		if !cached {
			return nil, fmt.Errorf("market %s does not exist", market)
		}
		cached = false
	}
}

func (self *Client) CreateOrder(order *Order) (*Order, error) {
	var (
		err       error
		unwrapped coinbasepro.Order
		wrapped   *Order
	)
	if unwrapped, err = self.Client.CreateOrder(order.Order); err != nil {
		return nil, err
	}
	if wrapped, err = wrap(&unwrapped); err != nil {
		return nil, err
	}
	return wrapped, nil
}

func ReadOnly() *Client {
	client := coinbasepro.NewClient()

	client.HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	if flag.Test() {
		client.UpdateConfig(&coinbasepro.ClientConfig{
			BaseURL: baseURL,
		})
	} else {
		client.UpdateConfig(&coinbasepro.ClientConfig{
			BaseURL: sandbox,
		})
	}

	return &Client{Client: client}
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

	apiPassphrase, err := flag.ApiPassphrase()
	if err != nil {
		return nil, err
	}

	client := ReadOnly()

	client.Client.UpdateConfig(&coinbasepro.ClientConfig{
		Key:        apiKey,
		Passphrase: apiPassphrase,
		Secret:     apiSecret,
	})

	return client, nil
}
