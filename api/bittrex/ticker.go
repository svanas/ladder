package bittrex

type Ticker struct {
	Symbol        string  `json:"symbol"`
	LastTradeRate float64 `json:"lastTradeRate,string"`
	BidRate       float64 `json:"bidRate,string"`
	AskRate       float64 `json:"askRate,string"`
}
