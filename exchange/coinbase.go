//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"github.com/svanas/ladder/api/coinbase"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
	"math/big"
	"strconv"
	"strings"
)

type Coinbase struct {
	*info
}

func (self *Coinbase) Cancel(market string, side consts.OrderSide) error {
	client, err := coinbase.New()
	if err != nil {
		return err
	}

	orders, err := client.GetOpenOrders(market, side)
	if err != nil {
		return err
	}

	var orderIds []string
	for _, order := range orders {
		orderIds = append(orderIds, order.OrderId)
	}

	return client.CancelOrders(orderIds)
}

func (self *Coinbase) FormatSymbol(asset string) (string, error) {
	return strings.ToUpper(asset), nil
}

func (self *Coinbase) FormatMarket(asset, quote string) (string, error) {
	return strings.ToUpper(fmt.Sprintf("%s-%s", asset, quote)), nil
}

func (self *Coinbase) Info() *info {
	return self.info
}

func (self *Coinbase) Nonce() (*big.Int, error) {
	return big.NewInt(0), nil
}

func (self *Coinbase) Order(market string, side consts.OrderSide, size, price big.Float, nonce big.Int, days int) error {
	client, err := coinbase.New()
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

func (self *Coinbase) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := coinbase.New()
	if err != nil {
		return nil, err
	}

	orders, err := client.GetOpenOrders(market, side)
	if err != nil {
		return nil, err
	}

	var output []Order
	for _, order := range orders {
		if order.Configuration.Limit.Size > 0 && order.Configuration.Limit.Price > 0 {
			output = append(output, Order{
				Size:  order.Configuration.Limit.Size,
				Price: order.Configuration.Limit.Price,
			})
		}
	}

	return output, nil
}

func (self *Coinbase) Precision(market string) (*Precision, error) {
	client, err := coinbase.New()
	if err != nil {
		return nil, err
	}
	product, err := client.GetProduct(market)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: precision.Parse(product.QuoteIncrement),
		Size:  precision.Parse(product.BaseIncrement),
	}, nil
}

func (self *Coinbase) Ticker(market string) (float64, error) {
	client, err := coinbase.New()
	if err != nil {
		return 0, err
	}
	product, err := client.GetProduct(market)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(product.Price, 64)
}

func newCoinbase() Exchange {
	return &Coinbase{
		info: &info{
			code: "COIN",
			name: "Coinbase",
		},
	}
}
