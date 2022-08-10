//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
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

type Precision struct {
	Price int
	Size  int
}

type Exchange interface {
	Cancel(market string, side consts.Side) error
	FormatMarket(asset, quote string) string
	Info() *info
	Order(side consts.Side, market string, size, price float64) (oid []byte, err error)
	Precision(market string) (*Precision, error)
	Ticker(market string) (float64, error)
}

var exchanges []Exchange

func init() {
	exchanges = append(exchanges, newCoinbasePro())
	exchanges = append(exchanges, newBitstamp())
}

func FindByName(name string) (Exchange, error) {
	for _, exchange := range exchanges {
		if exchange.Info().equals(name) {
			return exchange, nil
		}
	}
	return nil, fmt.Errorf("exchange %v does not exist", name)
}
