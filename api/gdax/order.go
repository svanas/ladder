//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package gdax

import (
	"encoding/json"
	coinbasepro "github.com/svanas/go-coinbasepro"
	"strconv"
)

type Order struct {
	*coinbasepro.Order
}

func (self *Order) GetSize() float64 {
	out, err := strconv.ParseFloat(self.Size, 64)
	if err == nil {
		return out
	}
	return 0
}

func (self *Order) SetSize(value float64) *Order {
	self.Size = strconv.FormatFloat(value, 'f', -1, 64)
	return self
}

func (self *Order) GetPrice() float64 {
	out, err := strconv.ParseFloat(self.Price, 64)
	if err == nil {
		return out
	}
	return 0
}

func (self *Order) SetPrice(value float64) *Order {
	self.Price = strconv.FormatFloat(value, 'f', -1, 64)
	return self
}

func wrap(input *coinbasepro.Order) (*Order, error) {
	var (
		err error
		buf []byte
		out Order
	)
	if buf, err = json.Marshal(input); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(buf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
