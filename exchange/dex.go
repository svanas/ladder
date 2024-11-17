package exchange

import (
	"fmt"
	"strings"

	"github.com/svanas/ladder/api/coingecko"
	"github.com/svanas/ladder/api/web3"
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

func (coin *coin) getDecimals(coingecko *coingecko.Client, chainId int64) (int, error) {
	if coin.id == "" {
		client, err := web3.New(chainId)
		if err != nil {
			return 0, err
		}
		return client.GetDecimals(coin.address)
	} else {
		return coingecko.GetDecimals(coin.id, chainId)
	}
}

func (dex *dex) parseMarket(chainId int64, market string) (*coin, *coin, error) { // --> (asset, quote, error)
	symbols := strings.Split(market, "-")
	if len(symbols) > 1 {
		assetId, _, assetAddr, err := dex.coingecko.GetCoin(symbols[0], chainId)
		if err != nil {
			if len(symbols[0]) == 42 && strings.HasPrefix(strings.ToLower(symbols[0]), "0x") {
				assetAddr = symbols[0]
			} else {
				return nil, nil, err
			}
		}
		quoteId, _, quoteAddr, err := dex.coingecko.GetCoin(symbols[1], chainId)
		if err != nil {
			if len(symbols[1]) == 42 && strings.HasPrefix(strings.ToLower(symbols[1]), "0x") {
				quoteAddr = symbols[1]
			} else {
				return nil, nil, err
			}
		}
		return &coin{assetId, assetAddr}, &coin{quoteId, quoteAddr}, nil
	}
	return nil, nil, fmt.Errorf("market %s does not exist", market)
}

func (dex *dex) formatSymbol(chainId int64, symbol string) (string, error) {
	_, sym, _, err := dex.coingecko.GetCoin(symbol, chainId)
	if err != nil {
		if len(symbol) == 42 && strings.HasPrefix(strings.ToLower(symbol), "0x") {
			client, err := web3.New(chainId)
			if err != nil {
				return "", err
			}
			if sym, err = client.GetSymbol(symbol); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return strings.ToUpper(sym), nil
}

func (dex *dex) precision(chainId int64, market string) (*Precision, error) {
	asset, quote, err := dex.parseMarket(chainId, market)
	if err != nil {
		return nil, err
	}
	assetDec, err := asset.getDecimals(dex.coingecko, chainId)
	if err != nil {
		return nil, err
	}
	quoteDec, err := quote.getDecimals(dex.coingecko, chainId)
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
	assetLast, err := func() (float64, error) {
		if asset.id == "" {
			return -1, nil
		} else {
			return dex.coingecko.GetTicker(asset.id)
		}
	}()
	if err != nil {
		return 0, err
	}
	quoteLast, err := func() (float64, error) {
		if quote.id == "" {
			return -1, nil
		} else {
			return dex.coingecko.GetTicker(quote.id)
		}
	}()
	if err != nil {
		return 0, err
	}
	return assetLast / quoteLast, nil
}
