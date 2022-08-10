//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package constants

import (
	"strings"
)

type Side string

const (
	BUY  Side = "buy"
	SELL Side = "sell"
)

func (self *Side) String() string {
	return string(*self)
}

func (self *Side) Equals(name string) bool {
	return strings.EqualFold(self.String(), name)
}
