//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package bitstamp

import (
	consts "github.com/svanas/ladder/constants"
)

type Order struct {
	Id           string  `json:"id"`
	DateTime     string  `json:"datetime"`
	Type         int     `json:"type,string"`
	Price        float64 `json:"price,string"`
	Amount       float64 `json:"amount,string"`
	CurrencyPair string  `json:"currency_pair,omitempty"` // warning: NOT equal to market name
}

func (self *Order) Side() consts.OrderSide {
	if self.Type == 0 {
		return consts.BUY
	}
	if self.Type == 1 {
		return consts.SELL
	}
	return ""
}
