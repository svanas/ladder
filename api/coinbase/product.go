//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package coinbase

import (
	"encoding/json"
)

type Product struct {
	ProductId                 string `json:"product_id"`                   // trading pair
	Price                     string `json:"price"`                        // current price for the product, in quote currency
	PricePercentageChange24h  string `json:"price_percentage_change_24h"`  // amount the price of the product has changed, in percent, in the last 24 hours
	Volume24h                 string `json:"volume_24h"`                   // trading volume for the product in the last 24 hours
	VolumePercentageChange24h string `json:"volume_percentage_change_24h"` // percentage amount the volume of the product has changed in the last 24 hours
	BaseIncrement             string `json:"base_increment"`               // minimum amount base value can be increased or decreased at once
	QuoteIncrement            string `json:"quote_increment"`              // minimum amount quote value can be increased or decreased at once
	QuoteMinSize              string `json:"quote_min_size"`               // minimum size that can be represented of quote currency
	QuoteMaxSize              string `json:"quote_max_size"`               // maximum size that can be represented of quote currency
	BaseMinSize               string `json:"base_min_size"`                // minimum size that can be represented of base currency
	BaseMaxSize               string `json:"base_max_size"`                // maximum size that can be represented of base currency
	BaseName                  string `json:"base_name"`                    // name of the base currency
	QuoteName                 string `json:"quote_name"`                   // name of the quote currency
	Watched                   bool   `json:"watched"`                      // whether or not the product is on the user's watchlist
	IsDisabled                bool   `json:"is_disabled"`                  // whether or not the product is disabled for trading
	New                       bool   `json:"new"`                          // whether or not the product is 'new'
	Status                    string `json:"status"`                       // status of the product
	CancelOnly                bool   `json:"cancel_only"`                  // whether or not orders of the product can only be cancelled, not placed or edited
	LimitOnly                 bool   `json:"limit_only"`                   // whether or not orders of the product can only be limit orders, not market orders
	PostOnly                  bool   `json:"post_only"`                    // whether or not orders of the product can only be posted, not cancelled
	TradingDisabled           bool   `json:"trading_disabled"`             // whether or not the product is disabled for trading for all market participants
	AuctionMode               bool   `json:"auction_mode"`                 // whether or not the product is in auction mode
	ProductType               string `json:"product_type"`                 // possible values are: [SPOT, FUTURE]
	QuoteCurrencyId           string `json:"quote_currency_id"`            // symbol of the quote currency
	BaseCurrencyId            string `json:"base_currency_id"`             // symbol of the base currency
	ViewOnly                  bool   `json:"view_only"`                    // whether or not the product is in view only mode
	PriceIncrement            string `json:"price_increment"`              // minimum amount price can be increased or decreased at once
}

func (self *Client) GetProducts() ([]Product, error) {
	data, err := self.get("products", nil)
	if err != nil {
		return nil, err
	}
	type Response struct {
		Products []Product `json:"products"`
	}
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return response.Products, nil
}

func (self *Client) GetProduct(productId string) (*Product, error) {
	data, err := self.get(("products/" + productId), nil)
	if err != nil {
		return nil, err
	}
	var product Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}
