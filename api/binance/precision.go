//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package binance

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/svanas/ladder/precision"
)

type (
	Prec struct {
		Symbol string
		Price  int
		Size   int
	}
	precs []Prec
)

func (self precs) indexBySymbol(symbol string) int {
	for i, p := range self {
		if p.Symbol == symbol {
			return i
		}
	}
	return -1
}

func (self precs) precFromSymbol(symbol string) *Prec {
	i := self.indexBySymbol(symbol)
	if i != -1 {
		return &self[i]
	}
	return nil
}

func getPrecsFromServer(client *binance.Client) (precs, error) {
	var output precs

	var info *binance.ExchangeInfo
	for {
		var err error
		info, err = func() (*binance.ExchangeInfo, error) {
			beforeRequest(client, exchangeInfo)
			defer afterRequest()
			return client.NewExchangeInfoService().Do(context.Background())
		}()
		if err == nil {
			break
		}
		if _, ok := handleRecvWindowError(client, err).(*errorContinue); !ok {
			return nil, err
		}
	}

	for _, symbol := range info.Symbols {
		prec := Prec{
			Symbol: symbol.Symbol,
		}
		for _, filter := range symbol.Filters {
			if filter["filterType"] == string(binance.SymbolFilterTypeLotSize) {
				if val, ok := filter["stepSize"]; ok {
					if str, ok := val.(string); ok {
						prec.Size = precision.Parse(str)
					}
				}
			}
			if filter["filterType"] == string(binance.SymbolFilterTypePriceFilter) {
				if val, ok := filter["tickSize"]; ok {
					if str, ok := val.(string); ok {
						prec.Price = precision.Parse(str)
					}
				}
			}
		}
		output = append(output, prec)
	}

	return output, nil
}

func getPrecs(client *binance.Client, cached bool) (precs, error) {
	if cache == nil || !cached {
		var err error
		if cache, err = getPrecsFromServer(client); err != nil {
			return nil, err
		}
	}
	return cache, nil
}

func (self *Client) GetPrec(symbol string) (*Prec, error) {
	cached := true
	for {
		precs, err := getPrecs(self.inner, cached)
		if err != nil {
			return nil, err
		}
		prec := precs.precFromSymbol(symbol)
		if prec != nil {
			return prec, nil
		}
		if !cached {
			return nil, fmt.Errorf("symbol %s does not exist", symbol)
		}
		cached = false
	}
}
