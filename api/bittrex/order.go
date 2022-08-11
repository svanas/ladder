package bittrex

type Order struct {
	Id            string  `json:"id"`
	MarketSymbol  string  `json:"marketSymbol"`
	Direction     string  `json:"direction"` // BUY or SELL
	OrderType     string  `json:"type"`      // LIMIT or MARKET
	Quantity      float64 `json:"quantity,string"`
	Limit         float64 `json:"limit,string,omitempty"`
	Ceiling       float64 `json:"ceiling,string,omitempty"`
	TimeInForce   string  `json:"timeInForce,omitempty"`
	ClientOrderId string  `json:"clientOrderId,omitempty"`
	FillQuantity  float64 `json:"fillQuantity,string,omitempty"`
	Commission    float64 `json:"commission,string,omitempty"`
	Proceeds      float64 `json:"proceeds,string,omitempty"`
	Status        string  `json:"status,omitempty"` // OPEN or CLOSED
	CreatedAt     string  `json:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
	ClosedAt      string  `json:"closedAt,omitempty"`
}
