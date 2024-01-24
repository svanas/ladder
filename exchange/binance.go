//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"strconv"
	"strings"

	"github.com/svanas/ladder/api/binance"
	consts "github.com/svanas/ladder/constants"
)

type Binance struct {
	*info
}

func (self *Binance) Cancel(market string, side consts.OrderSide) error {
	client, err := binance.ReadWrite()
	if err != nil {
		return err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if side.Equals(string(order.Side)) {
			if err := client.CancelOrder(market, order.OrderID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (self *Binance) FormatMarket(asset, quote string) (string, error) {
	return strings.ToUpper(asset + quote), nil
}

func (self *Binance) Info() *info {
	return self.info
}

func (self *Binance) Order(market string, side consts.OrderSide, size, price float64) (oid string, err error) {
	client, err := binance.ReadWrite()
	if err != nil {
		return "", err
	}

	order, err := client.CreateOrder(market, side, size, price)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(order.OrderID, 10), nil
}

func (self *Binance) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := binance.ReadWrite()
	if err != nil {
		return nil, err
	}

	orders, err := client.GetOpenOrders(market)
	if err != nil {
		return nil, err
	}

	parse := func(input string) float64 {
		out, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return out
		}
		return 0
	}

	var output []Order
	for _, order := range orders {
		if side.Equals(string(order.Side)) {
			output = append(output, Order{
				Size:  parse(order.OrigQuantity),
				Price: parse(order.Price),
			})
		}
	}

	return output, nil
}

func (self *Binance) Precision(symbol string) (*Precision, error) {
	client, err := binance.ReadOnly()
	if err != nil {
		return nil, err
	}
	prec, err := client.GetPrec(symbol)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: prec.Price,
		Size:  prec.Size,
	}, nil
}

func (self *Binance) Ticker(market string) (float64, error) {
	client, err := binance.ReadOnly()
	if err != nil {
		return 0, err
	}
	return client.GetTicker(market)
}

func newBinance() Exchange {
	return &Binance{
		info: &info{
			code: "BINA",
			name: "Binance",
		},
	}
}
