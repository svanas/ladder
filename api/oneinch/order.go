package oneinch

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/1inch/1inch-sdk/golang/client/orderbook"
)

func GetMakerAmount(order orderbook.OrderResponse) (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.Data.MakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.Data.MakingAmount)
	}
	return i, nil
}

func GetTakerAmount(order orderbook.OrderResponse) (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.Data.TakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.Data.TakingAmount)
	}
	return i, nil
}

func (client *Client) GetOrders(owner string) ([]orderbook.OrderResponse, error) {
	oneInchClient, err := client.oneInchClient()
	if err != nil {
		return nil, err
	}

	var (
		page   float32 = 1
		limit  float32 = 100
		output []orderbook.OrderResponse
	)
	for {
		orders, _, err := func() ([]orderbook.OrderResponse, *http.Response, error) {
			beforeRequest()
			defer afterRequest()
			return oneInchClient.Orderbook.GetOrdersByCreatorAddress(context.Background(), orderbook.GetOrdersByCreatorAddressParams{
				ChainId:        int(client.ChainId),
				CreatorAddress: owner,
				LimitOrderV3SubscribedApiControllerGetAllLimitOrdersParams: orderbook.LimitOrderV3SubscribedApiControllerGetAllLimitOrdersParams{
					Page:   page,
					Limit:  limit,
					SortBy: "createDateTime",
				},
			})
		}()
		if err != nil {
			return nil, err
		}
		output = append(output, orders...)
		if len(orders) < int(limit) {
			break
		}
		page++
	}

	return output, nil
}

func (client *Client) PlaceOrder(params orderbook.CreateOrderParams) error {

	// post the limit order
	oneInchClient, err := client.oneInchClient()
	if err != nil {
		return err
	}
	beforeRequest()
	defer afterRequest()

	_, _, err = oneInchClient.Orderbook.CreateOrder(context.Background(), params)
	if err != nil {
		if strings.Contains(err.Error(), "1inch router does not have approval for this token") {
			return fmt.Errorf("please approve %s on https://app.1inch.io/#/%d/advanced/limit-order", func() string {
				return params.MakerAsset
			}(), params.ChainId)
		} else {
			return err
		}
	}
	return nil
}
