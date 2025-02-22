//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package coinbase

import (
	"encoding/json"
	"errors"
	"net/url"

	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/uuid"
)

type Order struct {
	OrderId       string `json:"order_id"`   // unique id for this order
	ProductId     string `json:"product_id"` // product this order was created for e.g. 'BTC-USD'
	UserId        string `json:"user_id"`    // id of the user owning this Order
	Configuration struct {
		Limit struct {
			Size     float64 `json:"base_size,string"`   // amount of base currency to spend on order
			Price    float64 `json:"limit_price,string"` // ceiling price for which the order should get filled
			PostOnly bool    `json:"post_only"`          // post only limit order
		} `json:"limit_limit_gtc"`
	} `json:"order_configuration"`
	Side          string `json:"side"`            // possible values are: [UNKNOWN_ORDER_SIDE, BUY, SELL]
	ClientOrderId string `json:"client_order_id"` // client specified ID of order
	Status        string `json:"status"`          // possible values are: [OPEN, FILLED, CANCELLED, EXPIRED, FAILED, UNKNOWN_ORDER_STATUS]
	TimeInForce   string `json:"time_in_force"`   // possible values are: [UNKNOWN_TIME_IN_FORCE, GOOD_UNTIL_DATE_TIME, GOOD_UNTIL_CANCELLED, IMMEDIATE_OR_CANCEL, FILL_OR_KILL]
}

func (self *Client) GetOpenOrders(market string, side consts.OrderSide) ([]Order, error) {
	values := url.Values{}
	values.Add("product_id", market)
	values.Add("order_status", "OPEN")
	values.Add("order_side", side.String())
	data, err := self.get("orders/historical/batch", &values)
	if err != nil {
		return nil, err
	}
	type Response struct {
		Orders []Order `json:"orders"`
	}
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return response.Orders, nil
}

func (self *Client) CreateOrder(market string, side consts.OrderSide, size, price float64) (string, error) { // --> (orderId, error)
	type Request struct {
		ClientOrderId string `json:"client_order_id"`
		ProductId     string `json:"product_id"`
		Side          string `json:"side"`
		Configuration struct {
			Limit struct {
				Size  float64 `json:"base_size,string"`
				Price float64 `json:"limit_price,string"`
			} `json:"limit_limit_gtc"`
		} `json:"order_configuration"`
	}

	request := Request{
		ClientOrderId: uuid.New().String(),
		ProductId:     market,
		Side:          side.String(),
	}
	request.Configuration.Limit.Size = size
	request.Configuration.Limit.Price = price

	body, err := json.Marshal(&request)
	if err != nil {
		return "", err
	}

	data, err := self.post("orders", body)
	if err != nil {
		return "", err
	}

	type Response struct {
		Success         bool `json:"success"`
		SuccessResponse struct {
			OrderId string `json:"order_id"`
		} `json:"success_response"`
		ErrorResponse struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		} `json:"error_response"`
	}
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return "", err
	}
	if !response.Success {
		if response.ErrorResponse.Message != "" {
			return "", errors.New(response.ErrorResponse.Message)
		} else {
			return "", errors.New(response.ErrorResponse.Error)
		}
	}

	return response.SuccessResponse.OrderId, nil
}

func (self *Client) CancelOrders(orderIds []string) error {
	type Request struct {
		OrderIds []string `json:"order_ids"`
	}
	request := Request{OrderIds: orderIds}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if _, err := self.post("orders/batch_cancel", body); err != nil {
		return err
	}

	return nil
}
