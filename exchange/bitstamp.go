//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"strings"

	"github.com/svanas/ladder/api/bitstamp"
	consts "github.com/svanas/ladder/constants"
)

type Bitstamp struct {
	*info
}

func (self *Bitstamp) Cancel(market string, side consts.Side) error {
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

func (self *Bitstamp) FormatMarket(asset, quote string) string {
	return strings.ToLower(asset + quote)
}

func (self *Bitstamp) Info() *info {
	return self.info
}

func (self *Bitstamp) Order(side consts.Side, market string, size, price float64) (oid []byte, err error) {
	client, err := bitstamp.ReadWrite()
	if err != nil {
		return nil, err
	}

	order, err := func() (*bitstamp.Order, error) {
		if side == consts.BUY {
			return client.BuyLimitOrder(market, size, price)
		} else if side == consts.SELL {
			return client.SellLimitOrder(market, size, price)
		}
		return nil, fmt.Errorf("unknown order side %v", side)
	}()
	if err != nil {
		return nil, err
	}

	return []byte(order.Id), nil
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
	return ticker.Last, nil
}

func newBitstamp() Exchange {
	return &Bitstamp{
		info: &info{
			code: "BITS",
			name: "Bitstamp",
		},
	}
}
