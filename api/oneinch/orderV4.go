package oneinch

import (
	"crypto/rand"
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

type OrderDataV4 struct {
	Salt         string `json:"salt"`         // the highest 96 bits represent salt, and the lowest 160 bit represent extension hash.
	MakerAsset   string `json:"makerAsset"`   // the maker’s asset address.
	TakerAsset   string `json:"takerAsset"`   // the taker’s asset address.
	Maker        string `json:"maker"`        // the maker’s address
	Receiver     string `json:"receiver"`     // the receiver’s address. the taker assets will be transferred to this address.
	MakingAmount string `json:"makingAmount"` // the amount of tokens maker will give
	TakingAmount string `json:"takingAmount"` // the amount of tokens maker wants to receive
	MakerTraits  string `json:"makerTraits"`  // limit order options, coded as bit flags into uint256 number.
}

func (order *OrderDataV4) GetMakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.MakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.MakingAmount)
	}
	return i, nil
}

func (order *OrderDataV4) GetTakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.TakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.TakingAmount)
	}
	return i, nil
}

type OrderV4 struct {
	Signature string      `json:"signature"`
	OrderHash string      `json:"orderHash"`
	Data      OrderDataV4 `json:"data"`
}

func (client *Client) GetOrdersV4() ([]OrderV4, error) {
	owner, err := client.publicAddress()
	if err != nil {
		return nil, err
	}

	var (
		page   int = 1
		limit  int = 100
		output []OrderV4
	)
	for {
		orders, err := func() ([]OrderV4, error) {
			body, err := client.get(fmt.Sprintf("/orderbook/v4.0/%d/address/%s?page=%d&limit=%d&sortBy=createDateTime", client.ChainId, owner, page, limit))
			if err != nil {
				return nil, err
			}
			var response []OrderV4
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

func (client *Client) PlaceOrderV4(makerAsset, takerAsset string, makerAmount, takerAmount big.Float, nonce big.Int) error {
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
	allowance, err := web3.GetAllowance(makerAsset, maker, apiRouterV4)
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

	// compute the salt. the highest 96 bits represent the salt, and the lowest 160 bits are supposed to be zero
	salt, err := func() (*big.Int, error) {
		// define the maximum value (2^96 - 1)
		max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(1))
		// generate a random big.Int within the range [0, 2^96 - 1]
		salt, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, err
		}
		// shift the big.Int left by 160 bits to set the lowest 160 bits to zero
		salt.Lsh(salt, 160)
		return salt, nil
	}()
	if err != nil {
		return err
	}

	orderData := OrderDataV4{
		Salt:         salt.String(),
		MakerAsset:   makerAsset,
		TakerAsset:   takerAsset,
		Maker:        maker,
		Receiver:     taker,
		MakingAmount: precision.F2S(makerAmount, 0),
		TakingAmount: precision.F2S(takerAmount, 0),
		MakerTraits:  newMakerTraits(nonce, time.Now().Add(time.Hour).Unix()).encode(),
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
				{Name: "makingAmount", Type: "uint256"},
				{Name: "takingAmount", Type: "uint256"},
				{Name: "makerTraits", Type: "uint256"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "1inch Aggregation Router",
			Version:           "6",
			ChainId:           math.NewHexOrDecimal256(client.ChainId),
			VerifyingContract: apiRouterV4,
		},
		Message: apitypes.TypedDataMessage{
			"salt":         orderData.Salt,
			"makerAsset":   orderData.MakerAsset,
			"takerAsset":   orderData.TakerAsset,
			"maker":        orderData.Maker,
			"receiver":     orderData.Receiver,
			"makingAmount": orderData.MakingAmount,
			"takingAmount": orderData.TakingAmount,
			"makerTraits":  orderData.MakerTraits,
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
	body, err := json.Marshal(&OrderV4{
		OrderHash: challengeHash.Hex(),
		Signature: fmt.Sprintf("0x%x", signature),
		Data:      orderData,
	})
	if err != nil {
		return err
	}

	// post the limit order
	if _, err := client.post(fmt.Sprintf("/orderbook/v4.0/%d", client.ChainId), body); err != nil {
		return err
	}

	return nil
}
