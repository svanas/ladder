//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"strconv"

	coinbasepro "github.com/svanas/go-coinbasepro"
	"github.com/svanas/ladder/api/gdax"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type CoinbasePro struct {
	*info
}

func (self *CoinbasePro) Cancel(market string, side consts.OrderSide) error {
	client, err := gdax.ReadWrite()
	if err != nil {
		return err
	}

	cursor := client.ListOrders(coinbasepro.ListOrdersParams{
		Status:    "open",
		ProductID: market,
	})

	for cursor.HasMore {
		var orders []coinbasepro.Order
		if err := cursor.NextPage(&orders); err != nil {
			return err
		}
		for _, order := range orders {
			if side.Equals(order.Side) {
				if err := client.CancelOrder(order.ID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (self *CoinbasePro) FormatMarket(asset, quote string) string {
	return fmt.Sprintf("%s-%s", asset, quote)
}

func (self *CoinbasePro) Info() *info {
	return self.info
}

func (self *CoinbasePro) Order(side consts.OrderSide, market string, size, price float64) (oid *string, err error) {
	client, err := gdax.ReadWrite()
	if err != nil {
		return nil, err
	}

	input := (&gdax.Order{
		Order: &coinbasepro.Order{
			Type:      "limit",
			Side:      side.ToLowerCase(),
			ProductID: market,
		},
	}).SetSize(size).SetPrice(price)

	output, err := client.CreateOrder(input)
	if err != nil {
		return nil, err
	}

	return &output.ID, nil
}

func (self *CoinbasePro) Precision(market string) (*Precision, error) {
	product, err := gdax.ReadOnly().GetProduct(market)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: precision.Parse(product.QuoteIncrement),
		Size:  precision.Parse(product.BaseIncrement),
	}, nil
}

func (self *CoinbasePro) Ticker(market string) (float64, error) {
	ticker, err := gdax.ReadOnly().GetTicker(market)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ticker.Price, 64)
}

func newCoinbasePro() Exchange {
	return &CoinbasePro{
		info: &info{
			code: "GDAX",
			name: "Coinbase Pro",
		},
	}
}
