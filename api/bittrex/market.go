package bittrex

import (
	"strings"
)

type Market struct {
	Symbol              string   `json:"symbol"`
	BaseCurrencySymbol  string   `json:"baseCurrencySymbol"`
	QuoteCurrencySymbol string   `json:"quoteCurrencySymbol"`
	MinTradeSize        float64  `json:"minTradeSize,string"`
	Precision           int      `json:"precision"`
	Status              string   `json:"status"`
	CreatedAt           string   `json:"createdAt"`
	Notice              string   `json:"notice,omitempty"`
	ProhibitedIn        []string `json:"prohibitedIn,omitempty"`
}

// true if this market is currently online (and not about to be removed), otherwise false.
func (market *Market) Active() bool {
	return market.Status != "OFFLINE" && !strings.Contains(market.Notice, "will be removed")
}
