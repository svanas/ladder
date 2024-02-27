package kraken

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/svanas/kraken-go-api-client"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	inner *krakenapi.KrakenAPI
}

func (client *Client) Ticker(market string) (float64, error) {
	result, err := client.inner.Ticker(market)
	if err != nil {
		return 0, err
	}
	for _, info := range *result {
		return strconv.ParseFloat(info.Close[0], 64)
	}
	return 0, fmt.Errorf("market %s does not exist", market)
}

func (client *Client) PairInfo(market string) (*krakenapi.AssetPairInfo, error) {
	result, err := client.inner.AssetPair(market)
	if err != nil {
		return nil, err
	}
	for _, info := range *result {
		return &info, nil
	}
	return nil, fmt.Errorf("market %s does not exist", market)
}

type Order struct {
	TxId  string
	Order krakenapi.Order
}

func (client *Client) OpenOrders(market string) ([]Order, error) {
	orders, err := client.inner.OpenOrders(make(map[string]string))
	if err != nil {
		return nil, err
	}

	var output []Order
	for txid, order := range orders.Open {
		if order.Status == "open" && order.Description.Pair == market {
			output = append(output, Order{
				TxId:  txid,
				Order: order,
			})
		}
	}

	return output, nil
}

func (client *Client) CreateOrder(market string, side consts.OrderSide, size, price float64) (string, error) { // --> (txid, error)
	result, err := client.inner.AddOrder(market, side.ToLowerCase(), "limit", strconv.FormatFloat(size, 'f', -1, 64), map[string]string{
		"price": strconv.FormatFloat(price, 'f', -1, 64),
	})
	if err != nil {
		return "", err
	}
	return result.TxId[0], nil
}

func (client *Client) CancelOrder(txid string) error {
	result, err := client.inner.CancelOrder(txid)
	if err != nil {
		return err
	}
	if result.Count == 0 && !result.Pending {
		return fmt.Errorf("cannot cancel order %s", txid)
	}
	return nil
}

func new(apiKey, apiSecret string) (*Client, error) {
	return &Client{inner: krakenapi.NewWithClient(
		apiKey,
		apiSecret,
		&http.Client{
			Timeout: 30 * time.Second,
		})}, nil
}

func ReadOnly() (*Client, error) {
	return new("", "")
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

	return new(apiKey, apiSecret)
}
