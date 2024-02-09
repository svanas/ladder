package oneinch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/svanas/1inch-sdk/golang/client/orderbook"
	"github.com/svanas/1inch-sdk/golang/helpers/consts/contracts"
	"github.com/svanas/ladder/api/web3"
)

func GetMakerAmount(order *orderbook.OrderResponse) (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.Data.MakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.Data.MakingAmount)
	}
	return i, nil
}

func GetTakerAmount(order *orderbook.OrderResponse) (*big.Int, error) {
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

func (client *Client) PlaceOrder(params *orderbook.CreateOrderParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	// get the allowance, exit early when the 1inch router hasn't been approved
	web3, err := web3.New(int64(params.ChainId))
	if err != nil {
		return err
	}
	router, err := contracts.Get1inchRouterFromChainId(params.ChainId)
	if err != nil {
		return err
	}
	allowance, err := web3.GetAllowance(params.MakerAsset, params.Maker, router)
	if err != nil {
		return err
	}
	makerAmount, ok := new(big.Int).SetString(params.MakingAmount, 10)
	if !ok {
		return fmt.Errorf("cannot convert %s to big.Int", params.MakingAmount)
	}
	if allowance.Cmp(makerAmount) < 0 {
		return fmt.Errorf("please approve %s on https://app.1inch.io/#/%d/advanced/limit-order", func() string {
			if symbol, err := web3.GetSymbol(params.MakerAsset); err == nil && symbol != "" {
				return symbol
			}
			return params.MakerAsset
		}(), params.ChainId)
	}

	// from params to limit order
	order, err := orderbook.CreateLimitOrder(*params)
	if err != nil {
		return err
	}
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// post the limit order
	oneInchClient, err := client.oneInchClient()
	if err != nil {
		return err
	}
	request, err := oneInchClient.NewRequest("POST", fmt.Sprintf("/orderbook/v3.0/%d", params.ChainId), body)
	if err != nil {
		return err
	}
	var response orderbook.CreateOrderResponse
	if _, err := oneInchClient.Do(context.Background(), request, &response); err != nil {
		return err
	}
	if !response.Success {
		return errors.New("an unknown error occured. your limit order did not get saved")
	}

	return nil
}
