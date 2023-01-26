//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package constants

import (
	"strings"
)

//----------------------- OrderSide -----------------------

type OrderSide string

const (
	NONE OrderSide = ""
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

func (self *OrderSide) String() string {
	return string(*self)
}

func (self *OrderSide) ToLowerCase() string {
	return strings.ToLower(self.String())
}

func (self *OrderSide) ToUpperCase() string {
	return strings.ToUpper(self.String())
}

func (self *OrderSide) Equals(name string) bool {
	return strings.EqualFold(self.String(), name)
}

//----------------------- OrderType -----------------------

type OrderType string

const (
	LIMIT  OrderType = "LIMIT"
	MARKET OrderType = "MARKET"
)

func (self *OrderType) String() string {
	return string(*self)
}

//---------------------- TimeInForce ----------------------

type TimeInForce string

const (
	GTC TimeInForce = "GOOD_TIL_CANCELLED"
	IOC TimeInForce = "IMMEDIATE_OR_CANCEL"
	FOK TimeInForce = "FILL_OR_KILL"
)

func (self *TimeInForce) String() string {
	return string(*self)
}
