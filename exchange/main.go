package exchange

import (
	"strings"
)

type info struct {
	code string
	name string
}

func (info *info) equals(name string) bool {
	return strings.EqualFold(info.code, name) || strings.EqualFold(info.name, name)
}

type Order struct {
	Price float64
	Size  float64
}

type Exchange interface {
	Info() *info
	Sell(cancel bool, market string, orders []Order) error
}

var exchanges []Exchange

func init() {
	exchanges = append(exchanges, newCoinbasePro())
}

func FindByName(name string) Exchange {
	for _, exchange := range exchanges {
		if exchange.Info().equals(name) {
			return exchange
		}
	}
	return nil
}
