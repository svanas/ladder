//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package exchange

type CoinbasePro struct {
	*info
}

func (self *CoinbasePro) Info() *info {
	return self.info
}

func (self *CoinbasePro) Sell(cancel bool, market string, orders []Order) error {
	return nil
}

func newCoinbasePro() Exchange {
	return &CoinbasePro{
		info: &info{
			code: "GDAX",
			name: "Coinbase Pro",
		},
	}
}
