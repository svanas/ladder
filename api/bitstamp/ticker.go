package bitstamp

type Ticker struct {
	Last      string `json:"last"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Vwap      string `json:"vwap"`
	Volume    string `json:"volume"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
}
