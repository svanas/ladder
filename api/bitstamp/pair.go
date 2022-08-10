package bitstamp

type Pair struct {
	BaseDecimals    int    `json:"base_decimals"`    // size precision
	MinimumOrder    string `json:"minimum_order"`    // minimum order size
	CounterDecimals int    `json:"counter_decimals"` // price precision
	Trading         string `json:"trading"`          // enabled/disabled
	UrlSymbol       string `json:"url_symbol"`       // name
	Description     string `json:"description"`
}
