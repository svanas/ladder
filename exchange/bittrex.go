//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"strings"

	"github.com/svanas/ladder/api/bittrex"
	consts "github.com/svanas/ladder/constants"
)

type Bittrex struct {
	*info
}

func (self *Bittrex) Cancel(market string, side consts.OrderSide) error {
	client, err := bittrex.ReadWrite()
	if err != nil {
		return err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if side.Equals(order.Direction) {
			if err := client.CancelOrder(order.Id); err != nil {
				return err
			}
		}
	}

	return nil
}

func (self *Bittrex) FormatMarket(asset, quote string) string {
	return strings.ToUpper(fmt.Sprintf("%s-%s", asset, quote))
}

func (self *Bittrex) Info() *info {
	return self.info
}

func (self *Bittrex) Order(market string, side consts.OrderSide, size, price float64) (oid string, err error) {
	client, err := bittrex.ReadWrite()
	if err != nil {
		return "", err
	}

	order, err := client.CreateOrder(market, side, consts.LIMIT, size, price, consts.GTC)
	if err != nil {
		return "", err
	}

	return order.Id, nil
}

func (self *Bittrex) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := bittrex.ReadWrite()
	if err != nil {
		return nil, err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return nil, err
	}

	var output []Order
	for _, order := range orders {
		if side.Equals(order.Direction) {
			output = append(output, Order{
				Size:  order.Quantity,
				Price: order.Limit,
			})
		}
	}

	return output, nil
}

func (self *Bittrex) Precision(symbol string) (*Precision, error) {
	market, err := bittrex.ReadOnly().GetMarket(symbol)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: market.Precision,
		Size:  8,
	}, nil
}

func (self *Bittrex) Ticker(market string) (float64, error) {
	ticker, err := bittrex.ReadOnly().GetTicker(market)
	if err != nil {
		return 0, err
	}
	return ticker.LastTradeRate, nil
}

func newBittrex() Exchange {
	return &Bittrex{
		info: &info{
			code: "BTRX",
			name: "Bittrex",
		},
	}
}
