package oneinch

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/svanas/ladder/precision"
)

type IntegratorFee struct {
	Integrator string
	Protocol   string
	Fee        int // fee in basis points (e.g. 1 = 0.01%, 100 = 1%)
	Share      int // integrator's share in basis points (e.g. 1 = 0.01%, 100 = 1%)
}

func getIntegratorFee() *IntegratorFee {
	return &IntegratorFee{
		Integrator: "0x0000000000000000000000000000000000000000",
		Protocol:   "0x0000000000000000000000000000000000000000",
		Fee:        0,
		Share:      0,
	}
}

type ResolverFee struct {
	Whitelist                map[string]string `json:"whitelist"`
	FeeBps                   int               `json:"feeBps"`                   // fee in basis points (e.g. 50 = 0.5%)
	WhitelistDiscountPercent int               `json:"whitelistDiscountPercent"` // discount percentage for whitelisted resolvers (e.g. 50 = 50% off)
	ProtocolFeeReceiver      string            `json:"protocolFeeReceiver"`
	ExtensionAddress         string            `json:"extensionAddress"`
}

func (client *Client) getFeeInfo(makerAsset, takerAsset string, makerAmount, takerAmount big.Float) (*ResolverFee, error) {
	body, err := client.get(fmt.Sprintf("/orderbook/v4.1/%d/fee-info?makerAsset=%s&takerAsset=%s&makerAmount=%s&takerAmount=%s",
		client.ChainId,
		makerAsset,
		takerAsset,
		precision.F2S(makerAmount, 0),
		precision.F2S(takerAmount, 0),
	))
	if err != nil {
		return nil, err
	}
	var response ResolverFee
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
