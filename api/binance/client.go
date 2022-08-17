package binance

import (
	"context"
	"net/http"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/svanas/ladder/flag"
)

type Client struct {
	inner *binance.Client
}

var (
	server_time_offset int64 // offset between device time and server time
	cache              precs
)

func New(apiKey, apiSecret string) (*Client, error) {
	binance.UseTestnet = flag.Test()

	client := binance.NewClient(apiKey, apiSecret)
	client.HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	if server_time_offset == 0 {
		beforeRequest(client, serverTime)
		defer afterRequest()
		offset, err := client.NewSetServerTimeService().Do(context.Background())
		if err != nil {
			return nil, err
		}
		server_time_offset = offset
	}

	client.TimeOffset = server_time_offset

	return &Client{inner: client}, nil
}

func ReadOnly() (*Client, error) {
	return New("", "")
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

	return New(apiKey, apiSecret)
}
