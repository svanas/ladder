package oneinch

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/svanas/ladder/api/web3"
	"github.com/svanas/ladder/precision"
)

type OrderDataV3 struct {
	Salt          string `json:"salt"`
	MakerAsset    string `json:"makerAsset"`
	TakerAsset    string `json:"takerAsset"`
	Maker         string `json:"maker"`
	Receiver      string `json:"receiver"`
	AllowedSender string `json:"allowedSender"`
	MakingAmount  string `json:"makingAmount"`
	TakingAmount  string `json:"takingAmount"`
	Offsets       string `json:"offsets"`
	Interactions  string `json:"interactions"`
}

func (order *OrderDataV3) GetMakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.MakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.MakingAmount)
	}
	return i, nil
}

func (order *OrderDataV3) GetTakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.TakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.TakingAmount)
	}
	return i, nil
}

type OrderV3 struct {
	Signature string      `json:"signature"`
	OrderHash string      `json:"orderHash"`
	Data      OrderDataV3 `json:"data"`
}

func (client *Client) GetOrdersV3() ([]OrderV3, error) {
	owner, err := client.publicAddress()
	if err != nil {
		return nil, err
	}

	var (
		page   int = 1
		limit  int = 100
		output []OrderV3
	)
	for {
		orders, err := func() ([]OrderV3, error) {
			body, err := client.get(fmt.Sprintf("/orderbook/v3.0/%d/address/%s?page=%d&limit=%d&sortBy=createDateTime", client.ChainId, owner, page, limit))
			if err != nil {
				return nil, err
			}
			var response []OrderV3
			if err := json.Unmarshal(body, &response); err != nil {
				return nil, err
			}
			return response, nil
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

func (client *Client) PlaceOrderV3(makerAsset, takerAsset string, makerAmount, takerAmount big.Float) error {
	const taker = "0x0000000000000000000000000000000000000000"
	maker, err := client.publicAddress()
	if err != nil {
		return err
	}

	// get the allowance, exit early when the 1inch router hasn't been approved
	web3, err := web3.New(client.ChainId)
	if err != nil {
		return err
	}
	allowance, err := web3.GetAllowance(makerAsset, maker, apiRouterV3)
	if err != nil {
		return err
	}
	if new(big.Float).SetInt(allowance).Cmp(&makerAmount) < 0 {
		return fmt.Errorf("please approve %s on https://app.1inch.io/#/%d/advanced/limit-order", func() string {
			if symbol, err := web3.GetSymbol(makerAsset); err == nil && symbol != "" {
				return symbol
			}
			return makerAsset
		}(), client.ChainId)
	}

	orderData := OrderDataV3{
		Salt:          fmt.Sprintf("%d", (time.Now().UnixNano() / int64(time.Millisecond))),
		MakerAsset:    makerAsset,
		TakerAsset:    takerAsset,
		Maker:         maker,
		Receiver:      taker,
		AllowedSender: taker,
		MakingAmount:  precision.F2S(makerAmount, 0),
		TakingAmount:  precision.F2S(takerAmount, 0),
		Offsets:       "0",
		Interactions:  "0x",
	}

	// construct the ERC-712 message
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": []apitypes.Type{
				{Name: "salt", Type: "uint256"},
				{Name: "makerAsset", Type: "address"},
				{Name: "takerAsset", Type: "address"},
				{Name: "maker", Type: "address"},
				{Name: "receiver", Type: "address"},
				{Name: "allowedSender", Type: "address"},
				{Name: "makingAmount", Type: "uint256"},
				{Name: "takingAmount", Type: "uint256"},
				{Name: "offsets", Type: "uint256"},
				{Name: "interactions", Type: "bytes"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "1inch Aggregation Router",
			Version:           "5",
			ChainId:           math.NewHexOrDecimal256(client.ChainId),
			VerifyingContract: apiRouterV3,
		},
		Message: apitypes.TypedDataMessage{
			"salt":          orderData.Salt,
			"makerAsset":    orderData.MakerAsset,
			"takerAsset":    orderData.TakerAsset,
			"maker":         orderData.Maker,
			"receiver":      orderData.Receiver,
			"allowedSender": orderData.AllowedSender,
			"makingAmount":  orderData.MakingAmount,
			"takingAmount":  orderData.TakingAmount,
			"offsets":       orderData.Offsets,
			"interactions":  []byte{},
		},
	}

	// hash the ERC-712 message
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return err
	}
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return err
	}

	// prepare the data for signing
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	challengeHash := crypto.Keccak256Hash(rawData)

	// sign the challenge hash
	privateKey, err := client.ecdsaPrivateKey()
	if err != nil {
		return err
	}
	signature, err := crypto.Sign(challengeHash.Bytes(), privateKey)
	if err != nil {
		return err
	}

	// add 27 to `v` value (last byte)
	signature[64] += 27

	// construct the limit order
	body, err := json.Marshal(&OrderV3{
		OrderHash: challengeHash.Hex(),
		Signature: fmt.Sprintf("0x%x", signature),
		Data:      orderData,
	})
	if err != nil {
		return err
	}

	// post the limit order
	if _, err := client.post(fmt.Sprintf("/orderbook/v3.0/%d", client.ChainId), body); err != nil {
		return err
	}

	return nil
}
