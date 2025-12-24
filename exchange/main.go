//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"math/big"
	"strings"

	consts "github.com/svanas/ladder/constants"
)

type info struct {
	code string
	name string
}

func (self *info) equals(name string) bool {
	return strings.EqualFold(self.code, name) || strings.EqualFold(self.name, name)
}

type Order struct {
	Size  float64
	Price float64
}

func (order *Order) BigSize() big.Float {
	return *new(big.Float).SetFloat64(order.Size)
}

func (order *Order) BigPrice() big.Float {
	return *new(big.Float).SetFloat64(order.Price)
}

type Precision struct {
	Price int
	Size  int
}

type Exchange interface {
	Cancel(market string, side consts.OrderSide) error
	FormatSymbol(asset string) (string, error)
	FormatMarket(asset, quote string) (string, error)
	Info() *info
	Order(market string, side consts.OrderSide, size, price big.Float, days int) error
	Orders(market string, side consts.OrderSide) ([]Order, error)
	Precision(market string) (*Precision, error)
	Ticker(market string) (float64, error)
}

var exchanges []Exchange

func init() {
	exchanges = append(exchanges, newCoinbase())
	exchanges = append(exchanges, newBitstamp())
	exchanges = append(exchanges, newBinance())
	exchanges = append(exchanges, newOneInch())
	exchanges = append(exchanges, newKraken())
}

func FindByName(name string) (Exchange, error) {
	for _, exchange := range exchanges {
		if exchange.Info().equals(name) {
			return exchange, nil
		}
	}
	return nil, fmt.Errorf("exchange %v is not supported at this time", name)
}
