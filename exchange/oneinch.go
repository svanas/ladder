//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/svanas/ladder/api/coingecko"
	"github.com/svanas/ladder/api/oneinch"
	"github.com/svanas/ladder/api/web3"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type OneInch struct {
	*dex
}

func (self *OneInch) Cancel(market string, side consts.OrderSide) error {
	orders, err := self.Orders(market, side)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		return nil
	}
	symbols := strings.Split(market, "-")
	if len(symbols) < 2 {
		return fmt.Errorf("market %s does not exist", market)
	}
	asset, quote, err := func() (string, string, error) {
		switch side {
		case consts.BUY:
			return symbols[1], symbols[0], nil
		case consts.SELL:
			return symbols[0], symbols[1], nil
		}
		return "", "", fmt.Errorf("unknown order side %v", side)
	}()
	if err != nil {
		return err
	}
	client, err := oneinch.ReadOnly()
	if err != nil {
		return err
	}
	return fmt.Errorf("please cancel your orders on https://app.1inch.io/#/%d/advanced/limit-order/%s/%s", client.ChainId, asset, quote)
}

func (self *OneInch) FormatSymbol(asset string) (string, error) {
	client, err := oneinch.ReadOnly()
	if err != nil {
		return "", err
	}
	return self.formatSymbol(client.ChainId, asset)
}

func (self *OneInch) FormatMarket(asset, quote string) (string, error) {
	return self.formatMarket(asset, quote)
}

func (self *OneInch) Info() *info {
	return self.info
}

func (self *OneInch) Order(market string, side consts.OrderSide, size, price big.Float) error {
	client, err := oneinch.ReadWrite()
	if err != nil {
		return err
	}

	asset, quote, err := self.parseMarket(client.ChainId, market)
	if err != nil {
		return err
	}

	assetDec, err := self.coingecko.GetDecimals(asset.id, client.ChainId)
	if err != nil {
		return err
	}
	quoteDec, err := self.coingecko.GetDecimals(quote.id, client.ChainId)
	if err != nil {
		return err
	}

	// multiply an unscaled amount by these numbers to get the (scaled, non-floating) amount
	assetMul := new(big.Float).SetFloat64(math.Pow(10, float64(assetDec)))
	quoteMul := new(big.Float).SetFloat64(math.Pow(10, float64(quoteDec)))

	assetAmount := new(big.Float).Mul(&size, assetMul)
	quoteAmount := new(big.Float).Mul(new(big.Float).Mul(&size, &price), quoteMul)

	return func() error {
		switch side {
		case consts.BUY:
			return client.PlaceOrder(web3.Checksum(quote.address), web3.Checksum(asset.address), *quoteAmount, *assetAmount)
		case consts.SELL:
			return client.PlaceOrder(web3.Checksum(asset.address), web3.Checksum(quote.address), *assetAmount, *quoteAmount)
		}
		return fmt.Errorf("unknown order side %v", side)
	}()
}

func (self *OneInch) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := oneinch.ReadWrite()
	if err != nil {
		return nil, err
	}

	asset, quote, err := self.parseMarket(client.ChainId, market)
	if err != nil {
		return nil, err
	}

	assetDec, err := self.coingecko.GetDecimals(asset.id, client.ChainId)
	if err != nil {
		return nil, err
	}
	quoteDec, err := self.coingecko.GetDecimals(quote.id, client.ChainId)
	if err != nil {
		return nil, err
	}

	// divide a (scaled, non-floating) amount by these numbers to get the unscaled amount
	assetDiv := new(big.Float).SetFloat64(math.Pow(10, float64(assetDec)))
	quoteDiv := new(big.Float).SetFloat64(math.Pow(10, float64(quoteDec)))

	orders, err := client.GetOrders()
	if err != nil {
		return nil, err
	}
	var result []Order
	for _, order := range orders {
		makerScaled, err := order.Data.GetMakerAmount()
		if err != nil {
			return nil, err
		}
		takerScaled, err := order.Data.GetTakerAmount()
		if err != nil {
			return nil, err
		}
		if side == consts.BUY && strings.EqualFold(order.Data.MakerAsset, quote.address) && strings.EqualFold(order.Data.TakerAsset, asset.address) {
			makerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(makerScaled), quoteDiv).Float64()
			takerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(takerScaled), assetDiv).Float64()
			result = append(result, Order{
				Size:  takerUnscaled,
				Price: precision.Round((makerUnscaled / takerUnscaled), quoteDec),
			})
		}
		if side == consts.SELL && strings.EqualFold(order.Data.MakerAsset, asset.address) && strings.EqualFold(order.Data.TakerAsset, quote.address) {
			makerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(makerScaled), assetDiv).Float64()
			takerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(takerScaled), quoteDiv).Float64()
			result = append(result, Order{
				Size:  makerUnscaled,
				Price: precision.Round((takerUnscaled / makerUnscaled), quoteDec),
			})
		}
	}
	return result, nil
}

func (self *OneInch) Precision(market string) (*Precision, error) {
	client, err := oneinch.ReadOnly()
	if err != nil {
		return nil, err
	}
	return self.precision(client.ChainId, market)
}

func (self *OneInch) Ticker(market string) (float64, error) {
	client, err := oneinch.ReadOnly()
	if err != nil {
		return 0, err
	}
	return self.ticker(client.ChainId, market)
}

func newOneInch() Exchange {
	return &OneInch{
		dex: &dex{
			info: &info{
				name: "1inch",
			},
			coingecko: coingecko.New(),
		},
	}
}
