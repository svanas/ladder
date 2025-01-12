package command

import (
	"github.com/spf13/cobra"
	"github.com/svanas/ladder/answer"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/exchange"
	"github.com/svanas/ladder/flag"
	"github.com/svanas/ladder/internal"
)

func init() {
	sellCommand.Flags().String(consts.FLAG_ASSET, "", "name of the asset you will want to sell")
	sellCommand.Flags().String(consts.FLAG_QUOTE, "", "name of the asset you will want to receive")

	sellCommand.Flags().Float64(consts.START_AT_PRICE, 0, "price where you will want to start selling at")
	sellCommand.Flags().Float64(consts.STOP_AT_PRICE, 0, "price where you will want to stop selling")
	sellCommand.Flags().Float64(consts.START_WITH_SIZE, 0, "size of your first sell order (in base asset)")

	sellCommand.Flags().Float64(consts.FLAG_MULT, 1.05, "multiplier that defines the number of orders and the distance between them")
	sellCommand.Flags().Float64(consts.FLAG_SIZE, 0, "the quantity you will want to sell (in base asset)")

	sellCommand.Flags().String(consts.FLAG_EXCHANGE, "", "name or code of the exchange")
	sellCommand.Flags().Bool(consts.FLAG_DRY_RUN, true, "display the output of the command without actually running it")
	sellCommand.Flags().Bool(consts.FLAG_CANCEL, true, "cancel existing limit orders, if any")

	rootCommand.AddCommand(&sellCommand)
}

var sellCommand = cobra.Command{
	Use:   "sell",
	Short: "sell your crypto asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		asset, err := flag.GetString(*cmd, consts.FLAG_ASSET)
		if err != nil {
			return err
		}

		quote, err := flag.GetString(*cmd, consts.FLAG_QUOTE)
		if err != nil {
			return err
		}

		start_at_price, err := flag.GetFloat64(*cmd, consts.START_AT_PRICE)
		if err != nil {
			return err
		}

		stop_at_price, err := flag.GetFloat64(*cmd, consts.STOP_AT_PRICE)
		if err != nil {
			return err
		}

		if start_at_price > stop_at_price {
			stop_at_price, start_at_price = start_at_price, stop_at_price
		}

		start_with_size, err := flag.GetFloat64(*cmd, consts.START_WITH_SIZE)
		if err != nil {
			return err
		}

		mult, err := flag.Mult(*cmd)
		if err != nil {
			return err
		}

		size, err := flag.GetFloat64(*cmd, consts.FLAG_SIZE)
		if err != nil {
			return err
		}

		steps := 2
		for internal.SimulateSell(start_with_size, mult, steps) < size {
			steps++
		}
		steps--

		exc, err := func() (exchange.Exchange, error) {
			exc, err := flag.GetString(*cmd, consts.FLAG_EXCHANGE)
			if err != nil {
				return nil, err
			}
			return exchange.FindByName(exc)
		}()
		if err != nil {
			return err
		}

		market, err := exc.FormatMarket(asset, quote)
		if err != nil {
			return err
		}

		prec, err := exc.Precision(market)
		if err != nil {
			return err
		}

		dry_run, err := cmd.Flags().GetBool(consts.FLAG_DRY_RUN)
		if err != nil {
			return err
		}

		if asset, err = exc.FormatSymbol(asset); err != nil {
			return err
		}
		if quote, err = exc.FormatSymbol(quote); err != nil {
			return err
		}

		if !dry_run {
			// cancel existing limit sell orders
			cancel, err := cmd.Flags().GetBool(consts.FLAG_CANCEL)
			if err != nil {
				return err
			}
			if cancel {
				if err := exc.Cancel(market, consts.SELL); err != nil {
					return err
				}
			}
			// place new limit sell orders
			var (
				all bool // yes to all
				num int  // result
			)
			nonce, err := exc.Nonce()
			if err != nil {
				return err
			}
			ticker, err := exc.Ticker(market)
			if err != nil {
				return err
			}
			orders := internal.Orders(start_at_price, stop_at_price, start_with_size, mult, size, steps, *prec)
			for _, order := range orders {
				if (ticker == -1) || (order.Price > ticker) {
					yes := all
					if !yes {
						a := internal.Prompt(order, func() string {
							market, _ := exc.FormatMarket(asset, quote)
							return market
						}())
						yes = a == answer.YES || a == answer.YES_TO_ALL
						all = all || a == answer.YES_TO_ALL
					}
					if yes {
						if err := exc.Order(market, consts.SELL, order.BigSize(), order.BigPrice(), *nonce); err != nil {
							return err
						}
						num++
					}
				}
			}
		}

		internal.Print(asset, quote, start_at_price, stop_at_price, start_with_size, mult, size, steps, *prec)

		return nil
	},
}
