//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/svanas/ladder/api/coingecko"
	"github.com/svanas/ladder/api/paraswap"
	"github.com/svanas/ladder/api/web3"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/precision"
)

type ParaSwap struct {
	*dex
}

func (self *ParaSwap) Cancel(market string, side consts.OrderSide) error {
	orders, err := self.Orders(market, side)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		return nil
	}
	return errors.New("please cancel your orders on https://app.paraswap.io/#/limit")
}

func (self *ParaSwap) FormatSymbol(asset string) (string, error) {
	client, err := paraswap.ReadOnly()
	if err != nil {
		return "", err
	}
	return self.formatSymbol(client.ChainId, asset)
}

func (self *ParaSwap) FormatMarket(asset, quote string) (string, error) {
	return self.formatMarket(asset, quote)
}

func (self *ParaSwap) Info() *info {
	return self.info
}

func (self *ParaSwap) Nonce() (*big.Int, error) {
	return big.NewInt(0), nil
}

func (self *ParaSwap) Order(market string, side consts.OrderSide, size, price big.Float, nonce big.Int) error {
	client, err := paraswap.ReadWrite()
	if err != nil {
		return err
	}

	asset, quote, err := self.parseMarket(client.ChainId, market)
	if err != nil {
		return err
	}

	assetDec, err := asset.getDecimals(self.coingecko, client.ChainId)
	if err != nil {
		return err
	}
	quoteDec, err := quote.getDecimals(self.coingecko, client.ChainId)
	if err != nil {
		return err
	}

	// multiply an unscaled amount by these numbers to get the (scaled, non-floating) amount
	assetMul := new(big.Float).SetFloat64(math.Pow(10, float64(assetDec)))
	quoteMul := new(big.Float).SetFloat64(math.Pow(10, float64(quoteDec)))

	assetAmount := new(big.Float).Mul(&size, assetMul)
	quoteAmount := new(big.Float).Mul(new(big.Float).Mul(&size, &price), quoteMul)

	maker, err := client.PublicAddress()
	if err != nil {
		return err
	}

	order, err := func() (*paraswap.Order, error) {
		result := paraswap.Order{
			Maker: web3.Checksum(maker),
			Taker: "0x0000000000000000000000000000000000000000",
		}
		switch side {
		case consts.BUY:
			result.MakerAsset = web3.Checksum(quote.address)
			result.TakerAsset = web3.Checksum(asset.address)
			result.MakerAmount = precision.F2S(*quoteAmount, 0)
			result.TakerAmount = precision.F2S(*assetAmount, 0)
			return &result, nil
		case consts.SELL:
			result.MakerAsset = web3.Checksum(asset.address)
			result.TakerAsset = web3.Checksum(quote.address)
			result.MakerAmount = precision.F2S(*assetAmount, 0)
			result.TakerAmount = precision.F2S(*quoteAmount, 0)
			return &result, nil
		}
		return nil, fmt.Errorf("unknown order side %v", side)
	}()
	if err != nil {
		return err
	}

	return client.PlaceOrder(order)
}

func (self *ParaSwap) Orders(market string, side consts.OrderSide) ([]Order, error) {
	client, err := paraswap.ReadWrite()
	if err != nil {
		return nil, err
	}

	asset, quote, err := self.parseMarket(client.ChainId, market)
	if err != nil {
		return nil, err
	}

	assetDec, err := asset.getDecimals(self.coingecko, client.ChainId)
	if err != nil {
		return nil, err
	}
	quoteDec, err := quote.getDecimals(self.coingecko, client.ChainId)
	if err != nil {
		return nil, err
	}

	// divide a (scaled, non-floating) amount by these numbers to get the unscaled amount
	assetDiv := new(big.Float).SetFloat64(math.Pow(10, float64(assetDec)))
	quoteDiv := new(big.Float).SetFloat64(math.Pow(10, float64(quoteDec)))

	owner, err := client.PublicAddress()
	if err != nil {
		return nil, err
	}

	orders, err := client.GetOrders(owner)
	if err != nil {
		return nil, err
	}
	var result []Order
	for _, order := range orders {
		if order.Type == paraswap.LIMIT && order.State == paraswap.PENDING {
			makerScaled, err := order.GetMakerAmount()
			if err != nil {
				return nil, err
			}
			takerScaled, err := order.GetTakerAmount()
			if err != nil {
				return nil, err
			}
			if side == consts.BUY && strings.EqualFold(order.MakerAsset, quote.address) && strings.EqualFold(order.TakerAsset, asset.address) {
				makerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(makerScaled), quoteDiv).Float64()
				takerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(takerScaled), assetDiv).Float64()
				result = append(result, Order{
					Size:  takerUnscaled,
					Price: precision.Round((makerUnscaled / takerUnscaled), quoteDec),
				})
			}
			if side == consts.SELL && strings.EqualFold(order.MakerAsset, asset.address) && strings.EqualFold(order.TakerAsset, quote.address) {
				makerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(makerScaled), assetDiv).Float64()
				takerUnscaled, _ := new(big.Float).Quo(new(big.Float).SetInt(takerScaled), quoteDiv).Float64()
				result = append(result, Order{
					Size:  makerUnscaled,
					Price: precision.Round((takerUnscaled / makerUnscaled), quoteDec),
				})
			}
		}
	}
	return result, nil
}

func (self *ParaSwap) Precision(market string) (*Precision, error) {
	client, err := paraswap.ReadOnly()
	if err != nil {
		return nil, err
	}
	return self.precision(client.ChainId, market)
}

func (self *ParaSwap) Ticker(market string) (float64, error) {
	client, err := paraswap.ReadOnly()
	if err != nil {
		return 0, err
	}
	return self.ticker(client.ChainId, market)
}

func newParaSwap() Exchange {
	return &ParaSwap{
		dex: &dex{
			info: &info{
				code: "PSP",
				name: "ParaSwap",
			},
			coingecko: coingecko.New(),
		},
	}
}
