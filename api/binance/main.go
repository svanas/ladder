//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package binance

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/adshao/go-binance/v2"
	consts "github.com/svanas/ladder/constants"
)

func (self *Client) GetTicker(symbol string) (float64, error) {
	var tickers []*binance.SymbolPrice
	for {
		var err error
		tickers, err = func() ([]*binance.SymbolPrice, error) {
			beforeRequest(*self.inner, tickerPrice)
			defer afterRequest()
			return self.inner.NewListPricesService().Symbol(symbol).Do(context.Background())
		}()
		if err == nil {
			break
		}
		if _, ok := handleRecvWindowError(self.inner, err).(*errorContinue); !ok {
			return 0, err
		}
	}
	return strconv.ParseFloat(tickers[0].Price, 64)
}

func (self *Client) GetOpenOrders(symbol string) ([]*binance.Order, error) {
	var orders []*binance.Order
	for {
		var err error
		orders, err = func() ([]*binance.Order, error) {
			beforeRequest(*self.inner, openOrders)
			defer afterRequest()
			return self.inner.NewListOpenOrdersService().Symbol(symbol).Do(context.Background())
		}()
		if err == nil {
			break
		}
		if _, ok := handleRecvWindowError(self.inner, err).(*errorContinue); !ok {
			return nil, err
		}
	}
	return orders, nil
}

func (self *Client) CancelOrder(symbol string, orderID int64) error {
	for {
		_, err := func() (*binance.CancelOrderResponse, error) {
			beforeRequest(*self.inner, cancelOrder)
			defer afterRequest()
			return self.inner.NewCancelOrderService().Symbol(symbol).OrderID(orderID).Do(context.Background())
		}()
		if err == nil {
			break
		}
		if _, ok := handleRecvWindowError(self.inner, err).(*errorContinue); !ok {
			return err
		}
	}
	return nil
}

func (self *Client) CreateOrder(symbol string, side consts.OrderSide, size, price float64) (*binance.CreateOrderResponse, error) {
	clientOrderId := func() string {
		const (
			MAX_LEN = 36
			BROKER  = "J6MCRYME"
		)
		out := fmt.Sprintf("x-%s-", BROKER)
		for len(out) < MAX_LEN {
			n, _ := rand.Int(rand.Reader, big.NewInt(10))
			out += n.String()
		}
		return out
	}()

	var order *binance.CreateOrderResponse
	for {
		var err error
		order, err = func() (*binance.CreateOrderResponse, error) {
			beforeRequest(*self.inner, createOrder)
			defer afterRequest()
			return self.inner.NewCreateOrderService().
				Symbol(symbol).
				Side(binance.SideType(side.ToUpperCase())).
				Type(binance.OrderTypeLimit).
				TimeInForce(binance.TimeInForceTypeGTC).
				Quantity(strconv.FormatFloat(size, 'f', -1, 64)).
				Price(strconv.FormatFloat(price, 'f', -1, 64)).
				NewClientOrderID(clientOrderId).
				Do(context.Background())
		}()
		if err == nil {
			break
		}
		if _, ok := handleRecvWindowError(self.inner, err).(*errorContinue); !ok {
			return nil, err
		}
	}

	return order, nil
}
