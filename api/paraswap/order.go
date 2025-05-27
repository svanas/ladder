package paraswap

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/svanas/ladder/api/web3"
)

type OrderType string

const (
	LIMIT OrderType = "LIMIT"
	P2P   OrderType = "P2P"
)

type OrderState string

const (
	PENDING   OrderState = "PENDING"
	FULFILLED OrderState = "FULFILLED"
	CANCELLED OrderState = "CANCELLED"
	EXPIRED   OrderState = "EXPIRED"
)

type Order struct {
	Expiry       int64      `json:"expiry"`
	NonceAndMeta string     `json:"nonceAndMeta"`
	Maker        string     `json:"maker"`
	Taker        string     `json:"taker"`
	MakerAsset   string     `json:"makerAsset"`
	TakerAsset   string     `json:"takerAsset"`
	MakerAmount  string     `json:"makerAmount"`
	TakerAmount  string     `json:"takerAmount"`
	Signature    string     `json:"signature"`
	Type         OrderType  `json:"type,omitempty"`
	State        OrderState `json:"state,omitempty"`
}

func (order *Order) GetMakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.MakerAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.MakerAmount)
	}
	return i, nil
}

func (order *Order) GetTakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.TakerAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.TakerAmount)
	}
	return i, nil
}

func (client *Client) GetOrders(owner string) ([]Order, error) {
	var (
		result  []Order = nil
		hasMore bool    = true
		offset  int     = 0
	)
	for hasMore {
		body, err := client.get(fmt.Sprintf("ft/orders/%d/maker/%s?offset=%d", client.ChainId, owner, offset))
		if err != nil {
			return nil, err
		}
		type Response struct {
			Orders  []Order `json:"orders"`
			HasMore bool    `json:"hasMore"`
		}
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		result = append(result, response.Orders...)
		hasMore = response.HasMore
		offset++
	}
	return result, nil
}

func (client *Client) PlaceOrder(order *Order) error {
	// get the allowance, exit early when the ParaSwap router hasn't been approved
	web3, err := web3.New(client.ChainId)
	if err != nil {
		return err
	}
	router, err := router(client.ChainId)
	if err != nil {
		return err
	}
	allowance, err := web3.GetAllowance(order.MakerAsset, order.Maker, router)
	if err != nil {
		return err
	}
	makerAmount, err := order.GetMakerAmount()
	if err != nil {
		return err
	}
	if allowance.Cmp(makerAmount) < 0 {
		return fmt.Errorf("please approve %s on https://app.velora.xyz/#/limit", func() string {
			if symbol, err := web3.GetSymbol(order.MakerAsset); err == nil && symbol != "" {
				return symbol
			}
			return order.MakerAsset
		}())
	}

	// calculate nonceAndMeta (taker address plus a random integer between 0 and 2 ^ 53 - 1 shifted 160 bits)
	taker, ok := new(big.Int).SetString(order.Taker, 0)
	if !ok {
		return fmt.Errorf("cannot set %s to big.Int", order.Taker)
	}
	order.NonceAndMeta = new(big.Int).Add(taker, new(big.Int).Lsh(func() *big.Int {
		n, _ := rand.Int(rand.Reader, new(big.Int).SetInt64((2 ^ 53 - 2)))
		n = new(big.Int).Add(n, new(big.Int).SetInt64(1))
		return n
	}(), 160)).Text(10)

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
				{Name: "expiry", Type: "int256"},
				{Name: "nonceAndMeta", Type: "string"},
				{Name: "maker", Type: "string"},
				{Name: "taker", Type: "string"},
				{Name: "makerAsset", Type: "string"},
				{Name: "takerAsset", Type: "string"},
				{Name: "makerAmount", Type: "string"},
				{Name: "takerAmount", Type: "string"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "AUGUSTUS RFQ",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(client.ChainId),
			VerifyingContract: router,
		},
		Message: apitypes.TypedDataMessage{
			"expiry":       math.NewHexOrDecimal256(order.Expiry),
			"nonceAndMeta": order.NonceAndMeta,
			"maker":        order.Maker,
			"taker":        order.Taker,
			"makerAsset":   order.MakerAsset,
			"takerAsset":   order.TakerAsset,
			"makerAmount":  order.MakerAmount,
			"takerAmount":  order.TakerAmount,
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

	// convert signature to hex string
	order.Signature = fmt.Sprintf("0x%x", signature)

	rawBody, err := json.Marshal(order)
	if err != nil {
		return err
	}

	if _, err := client.post(fmt.Sprintf("ft/orders/%d", client.ChainId), rawBody); err != nil {
		return err
	}

	return nil
}
