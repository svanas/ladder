//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"errors"
	"fmt"
	"strings"

	"github.com/svanas/ladder/api/coinbase"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type Coinbase struct {
	*info
}

func (self *Coinbase) Cancel(market string, side consts.OrderSide) error {
	return errors.New("not implemented")
}

func (self *Coinbase) FormatMarket(asset, quote string) string {
	return strings.ToUpper(fmt.Sprintf("%s-%s", asset, quote))
}

func (self *Coinbase) Info() *info {
	return self.info
}

func (self *Coinbase) Order(market string, side consts.OrderSide, size, price float64) (oid string, err error) {
	return "", errors.New("not implemented")
}

func (self *Coinbase) Orders(market string, side consts.OrderSide) ([]Order, error) {
	return nil, errors.New("not implemented")
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
	return 0, errors.New("not implemented")
}

func newCoinbase() Exchange {
	return &Coinbase{
		info: &info{
			code: "COIN",
			name: "Coinbase",
		},
	}
}
