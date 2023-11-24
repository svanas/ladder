//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"errors"
	"fmt"
	"strings"

	consts "github.com/svanas/ladder/constants"
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

func (self *Coinbase) Precision(symbol string) (*Precision, error) {
	return nil, errors.New("not implemented")
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
