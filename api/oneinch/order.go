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
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type OrderData struct {
	Salt         string `json:"salt"`         // the highest 96 bits represent salt, and the lowest 160 bit represent extension hash.
	Maker        string `json:"maker"`        // the maker’s address
	Receiver     string `json:"receiver"`     // the receiver’s address. the taker assets will be transferred to this address.
	MakerAsset   string `json:"makerAsset"`   // the maker’s asset address.
	TakerAsset   string `json:"takerAsset"`   // the taker’s asset address.
	MakingAmount string `json:"makingAmount"` // the amount of tokens maker will give
	TakingAmount string `json:"takingAmount"` // the amount of tokens maker wants to receive
	MakerTraits  string `json:"makerTraits"`  // limit order options, coded as bit flags into uint256 number.
	Extension    string `json:"extension"`    // extensions are features that consume more gas to execute, but are not always necessary for a limit order.
}

func (order *OrderData) GetMakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.MakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.MakingAmount)
	}
	return i, nil
}

func (order *OrderData) GetTakerAmount() (*big.Int, error) {
	i, ok := new(big.Int).SetString(order.TakingAmount, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", order.TakingAmount)
	}
	return i, nil
}

type Order struct {
	Signature string    `json:"signature"`
	OrderHash string    `json:"orderHash"`
	Data      OrderData `json:"data"`
}

func (client *Client) GetOrders() ([]Order, error) {
	owner, err := client.publicAddress()
	if err != nil {
		return nil, err
	}

	var (
		page   int = 1
		limit  int = 100
		output []Order
	)
	for {
		orders, err := func() ([]Order, error) {
			body, err := client.get(fmt.Sprintf("/orderbook/v4.0/%d/address/%s?page=%d&limit=%d&sortBy=createDateTime", client.ChainId, owner, page, limit))
			if err != nil {
				return nil, err
			}
			var response []Order
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

func (client *Client) PlaceOrder(makerAsset, takerAsset string, makerAmount, takerAmount big.Float, nonce big.Int, days int) error {
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
	allowance, err := web3.GetAllowance(makerAsset, maker, apiRouter)
	if err != nil {
		return err
	}
	if new(big.Float).SetInt(allowance).Cmp(&makerAmount) < 0 {
		return fmt.Errorf("please approve %s on https://1inch.com/pro?mode=limit&pair=%d:%s-%s", func() string {
			if symbol, err := web3.GetSymbol(makerAsset); err == nil && symbol != "" {
				return symbol
			}
			return makerAsset
		}(), client.ChainId, makerAsset, takerAsset)
	}

	// compute the salt. the highest 96 bits represent salt, and the lowest 160 bit represent extension hash
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

	expiry := func() time.Duration {
		if days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
		return consts.THREE_YEARS
	}()

	orderData := OrderData{
		Salt:         salt.String(),
		Maker:        maker,
		Receiver:     taker,
		MakerAsset:   makerAsset,
		TakerAsset:   takerAsset,
		MakingAmount: precision.F2S(makerAmount, 0),
		TakingAmount: precision.F2S(takerAmount, 0),
		MakerTraits:  newMakerTraits(nonce, time.Now().Add(expiry).Unix()).encode(),
		Extension:    "0x",
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
				{Name: "maker", Type: "address"},
				{Name: "receiver", Type: "address"},
				{Name: "makerAsset", Type: "address"},
				{Name: "takerAsset", Type: "address"},
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
			VerifyingContract: apiRouter,
		},
		Message: apitypes.TypedDataMessage{
			"salt":         orderData.Salt,
			"maker":        orderData.Maker,
			"receiver":     orderData.Receiver,
			"makerAsset":   orderData.MakerAsset,
			"takerAsset":   orderData.TakerAsset,
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
	body, err := json.Marshal(&Order{
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
