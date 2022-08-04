package exchange

import (
	"fmt"
	"strings"
)

type info struct {
	code string
	name string
}

func (info *info) equals(name string) bool {
	return strings.EqualFold(info.code, name) || strings.EqualFold(info.name, name)
}

type Precision struct {
	Price int
	Size  int
}

type Side string

const (
	BUY  Side = "buy"
	SELL Side = "sell"
)

type Exchange interface {
	Cancel(market string, side Side) error
	FormatMarket(asset, quote string) string
	Info() *info
	Order(side Side, market string, size, price float64) (oid []byte, err error)
	Precision(market string) (*Precision, error)
	Ticker(market string) (float64, error)
}

var exchanges []Exchange

func init() {
	exchanges = append(exchanges, newCoinbasePro())
}

func FindByName(name string) (Exchange, error) {
	for _, exchange := range exchanges {
		if exchange.Info().equals(name) {
			return exchange, nil
		}
	}
	return nil, fmt.Errorf("exchange %v does not exist", name)
}
