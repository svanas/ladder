package binance

type request int

const (
	cancelOrder request = iota
	createOrder
	exchangeInfo
	openOrders
	serverTime
	tickerPrice
)

var weight = map[request]int{
	cancelOrder:  1,
	createOrder:  1,
	exchangeInfo: 10,
	openOrders:   3,
	serverTime:   1,
	tickerPrice:  1,
}
