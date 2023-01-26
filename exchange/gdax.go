//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"strconv"

	coinbase "github.com/svanas/go-coinbasepro"
	"github.com/svanas/ladder/api/gdax"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type Coinbase struct {
	*info
}

func (self *Coinbase) Cancel(market string, side consts.OrderSide) error {
	client, err := gdax.ReadWrite()
	if err != nil {
		return err
	}

	cursor := client.ListOrders(coinbase.ListOrdersParams{
		Status:    "open",
		ProductID: market,
	})

	for cursor.HasMore {
		var orders []coinbase.Order
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

func (self *Coinbase) FormatMarket(asset, quote string) string {
	return fmt.Sprintf("%s-%s", asset, quote)
}

func (self *Coinbase) Info() *info {
	return self.info
}

func (self *Coinbase) Order(market string, side consts.OrderSide, size, price float64) (oid string, err error) {
	client, err := gdax.ReadWrite()
	if err != nil {
		return "", err
	}

	input := (&gdax.Order{
		Order: &coinbase.Order{
			Type:      "limit",
			Side:      side.ToLowerCase(),
			ProductID: market,
		},
	}).SetSize(size).SetPrice(price)

	output, err := client.CreateOrder(input)
	if err != nil {
		return "", err
	}

	return output.ID, nil
}

func (self *Coinbase) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := gdax.ReadWrite()
	if err != nil {
		return nil, err
	}

	cursor := client.ListOrders(coinbase.ListOrdersParams{
		Status:    "open",
		ProductID: market,
	})

	var output []Order
	for cursor.HasMore {
		var orders []coinbase.Order
		if err := cursor.NextPage(&orders); err != nil {
			return nil, err
		}
		for _, order := range orders {
			if side.Equals(order.Side) {
				wrapped, err := gdax.Wrap(&order)
				if err != nil {
					return nil, err
				}
				output = append(output, Order{
					Size:  wrapped.GetSize(),
					Price: wrapped.GetPrice(),
				})
			}
		}
	}

	return output, nil
}

func (self *Coinbase) Precision(market string) (*Precision, error) {
	product, err := gdax.ReadOnly().GetProduct(market)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: precision.Parse(product.QuoteIncrement),
		Size:  precision.Parse(product.BaseIncrement),
	}, nil
}

func (self *Coinbase) Ticker(market string) (float64, error) {
	ticker, err := gdax.ReadOnly().GetTicker(market)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ticker.Price, 64)
}

func newCoinbase() Exchange {
	return &Coinbase{
		info: &info{
			code: "GDAX",
			name: "Coinbase",
		},
	}
}
