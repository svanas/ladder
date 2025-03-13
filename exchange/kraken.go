//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"math/big"
	"strings"

	"github.com/svanas/ladder/api/kraken"
	consts "github.com/svanas/ladder/constants"
)

type Kraken struct {
	*info
}

func (_ *Kraken) Cancel(market string, side consts.OrderSide) error {
	client, err := kraken.ReadWrite()
	if err != nil {
		return err
	}

	orders, err := client.OpenOrders(market)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if side.Equals(order.Order.Description.Type) && order.Order.Description.OrderType == "limit" {
			if err := client.CancelOrder(order.TxId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (_ *Kraken) FormatSymbol(asset string) (string, error) {
	return strings.ToUpper(asset), nil
}

func (_ *Kraken) FormatMarket(asset, quote string) (string, error) {
	return strings.ToUpper(asset + quote), nil
}

func (self *Kraken) Info() *info {
	return self.info
}

func (_ *Kraken) Order(market string, side consts.OrderSide, size, price big.Float, days int) error {
	client, err := kraken.ReadWrite()
	if err != nil {
		return err
	}
	if _, err := client.CreateOrder(market, side, func() float64 {
		out, _ := size.Float64()
		return out
	}(), func() float64 {
		out, _ := price.Float64()
		return out
	}()); err != nil {
		return err
	}
	return nil
}

func (_ *Kraken) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := kraken.ReadWrite()
	if err != nil {
		return nil, err
	}

	orders, err := client.OpenOrders(market)
	if err != nil {
		return nil, err
	}

	var output []Order
	for _, order := range orders {
		if side.Equals(order.Order.Description.Type) && order.Order.Description.OrderType == "limit" && order.Order.Description.Price > 0 && order.Order.Volume > 0 {
			output = append(output, Order{
				Size:  order.Order.Volume,
				Price: order.Order.Description.Price,
			})
		}
	}

	return output, nil
}

func (_ *Kraken) Precision(market string) (*Precision, error) {
	client, err := kraken.ReadOnly()
	if err != nil {
		return nil, err
	}
	info, err := client.PairInfo(market)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: info.PairDecimals,
		Size:  info.CostDecimals,
	}, nil
}

func (_ *Kraken) Ticker(market string) (float64, error) {
	client, err := kraken.ReadOnly()
	if err != nil {
		return 0, err
	}
	return client.Ticker(market)
}

func newKraken() Exchange {
	return &Kraken{
		info: &info{
			code: "KRKN",
			name: "Kraken",
		},
	}
}
