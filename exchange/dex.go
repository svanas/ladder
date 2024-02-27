package exchange

import (
	"fmt"
	"strings"

	"github.com/svanas/ladder/api/coingecko"
)

// abstract base struct for DEXes
type dex struct {
	*info
	coingecko *coingecko.Client
}

func (dex *dex) formatMarket(asset, quote string) (string, error) {
	return strings.ToUpper(fmt.Sprintf("%s-%s", asset, quote)), nil
}

type coin struct {
	id      string // coingecko coin id
	address string // on-chain token address
}

func (dex *dex) parseMarket(chainId int64, market string) (*coin, *coin, error) { // --> (asset, quote, error)
	symbols := strings.Split(market, "-")
	if len(symbols) > 1 {
		assetId, _, assetAddr, err := dex.coingecko.GetCoin(symbols[0], chainId)
		if err != nil {
			return nil, nil, err
		}
		quoteId, _, quoteAddr, err := dex.coingecko.GetCoin(symbols[1], chainId)
		if err != nil {
			return nil, nil, err
		}
		return &coin{assetId, assetAddr}, &coin{quoteId, quoteAddr}, nil
	}
	return nil, nil, fmt.Errorf("market %s does not exist", market)
}

func (dex *dex) formatSymbol(chainId int64, symbol string) (string, error) {
	_, sym, _, err := dex.coingecko.GetCoin(symbol, chainId)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(sym), nil
}

func (dex *dex) precision(chainId int64, market string) (*Precision, error) {
	asset, quote, err := dex.parseMarket(chainId, market)
	if err != nil {
		return nil, err
	}
	assetDec, err := dex.coingecko.GetDecimals(asset.id, chainId)
	if err != nil {
		return nil, err
	}
	quoteDec, err := dex.coingecko.GetDecimals(quote.id, chainId)
	if err != nil {
		return nil, err
	}
	return &Precision{
		Price: quoteDec,
		Size:  assetDec,
	}, nil
}

func (dex *dex) ticker(chainId int64, market string) (float64, error) {
	asset, quote, err := dex.parseMarket(chainId, market)
	if err != nil {
		return 0, err
	}
	assetLast, err := dex.coingecko.GetTicker(asset.id)
	if err != nil {
		return 0, err
	}
	quoteLast, err := dex.coingecko.GetTicker(quote.id)
	if err != nil {
		return 0, err
	}
	return assetLast / quoteLast, nil
}
