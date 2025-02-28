//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/svanas/ladder/api/bitstamp"
	consts "github.com/svanas/ladder/constants"
)

type Bitstamp struct {
	*info
}

func (self *Bitstamp) Cancel(market string, side consts.OrderSide) error {
	client, err := bitstamp.ReadWrite()
	if err != nil {
		return err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if order.Side() == side {
			if err := client.CancelOrder(order.Id); err != nil {
				return err
			}
		}
	}

	return nil
}

func (self *Bitstamp) FormatSymbol(asset string) (string, error) {
	return strings.ToLower(asset), nil
}

func (self *Bitstamp) FormatMarket(asset, quote string) (string, error) {
	return strings.ToLower(asset + quote), nil
}

func (self *Bitstamp) Info() *info {
	return self.info
}

func (self *Bitstamp) Nonce() (*big.Int, error) {
	return big.NewInt(0), nil
}

func (self *Bitstamp) Order(market string, side consts.OrderSide, size, price big.Float, nonce big.Int, days int) error {
	client, err := bitstamp.ReadWrite()
	if err != nil {
		return err
	}

	s, _ := size.Float64()
	p, _ := price.Float64()

	if _, err := func() (*bitstamp.Order, error) {
		if side == consts.BUY {
			return client.BuyLimitOrder(market, s, p)
		} else if side == consts.SELL {
			return client.SellLimitOrder(market, s, p)
		}
		return nil, fmt.Errorf("unknown order side %v", side)
	}(); err != nil {
		return err
	}

	return nil
}

func (self *Bitstamp) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := bitstamp.ReadWrite()
	if err != nil {
		return nil, err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return nil, err
	}

	var output []Order
	for _, order := range orders {
		if order.Side() == side {
			output = append(output, Order{
				Size:  order.Amount,
				Price: order.Price,
			})
		}
	}

	return output, nil
}

func (self *Bitstamp) Precision(market string) (*Precision, error) {
	pair, err := bitstamp.ReadOnly().GetPair(market)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: pair.CounterDecimals,
		Size:  pair.BaseDecimals,
	}, nil
}

func (self *Bitstamp) Ticker(market string) (float64, error) {
	ticker, err := bitstamp.ReadOnly().Ticker(market)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ticker.Last, 64)
}

func newBitstamp() Exchange {
	return &Bitstamp{
		info: &info{
			code: "BITS",
			name: "Bitstamp",
		},
	}
}
