package command

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	consts "github.com/svanas/ladder/constants"
	"github.com/svanas/ladder/exchange"
	"github.com/svanas/ladder/flag"
)

func init() {
	cancelCommand.Flags().String(consts.FLAG_ASSET, "BTC", "base asset")
	cancelCommand.Flags().String(consts.FLAG_QUOTE, "USDT", "quote asset")

	cancelCommand.Flags().String(consts.FLAG_EXCHANGE, "", "name or code of the exchange")
	cancelCommand.Flags().Bool(consts.FLAG_DRY_RUN, true, "display the output of the command without actually running it")

	cancelCommand.Flags().String(consts.FLAG_SIDE, "", "\"buy\" or \"sell\"")

	rootCommand.AddCommand(cancelCommand)
}

var cancelCommand = &cobra.Command{
	Use:   "cancel",
	Short: "cancel your open orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		asset, err := cmd.Flags().GetString(consts.FLAG_ASSET)
		if err != nil {
			return err
		}

		quote, err := cmd.Flags().GetString(consts.FLAG_QUOTE)
		if err != nil {
			return err
		}

		cex, err := func() (exchange.Exchange, error) {
			cex, err := flag.GetString(cmd, consts.FLAG_EXCHANGE)
			if err != nil {
				return nil, err
			}
			return exchange.FindByName(cex)
		}()
		if err != nil {
			return err
		}

		side, err := func() (consts.OrderSide, error) {
			side, err := flag.GetString(cmd, consts.FLAG_SIDE)
			if err != nil {
				return consts.NONE, err
			} else if side == "buy" {
				return consts.BUY, nil
			} else if side == "sell" {
				return consts.SELL, nil
			} else {
				return consts.NONE, fmt.Errorf("--%s is invalid. valid values are \"buy\" or \"sell\"", consts.FLAG_SIDE)
			}
		}()
		if err != nil {
			return err
		}

		market := cex.FormatMarket(asset, quote)

		prec, err := cex.Precision(market)
		if err != nil {
			return err
		}

		dry_run, err := cmd.Flags().GetBool(consts.FLAG_DRY_RUN)
		if err != nil {
			return err
		}

		if !dry_run {
			if err := cex.Cancel(market, side); err != nil {
				return err
			}
		} else {
			orders, err := cex.Orders(market, side)
			if err != nil {
				return err
			}

			writer := table.NewWriter()
			writer.AppendHeader(table.Row{"", "Side", "Price", "Size", "Value"})

			for index, order := range orders {
				writer.AppendRow(table.Row{index + 1, side.String(),
					fmt.Sprintf("%[3]v %.[2]*[1]f", order.Price, prec.Price, quote),
					fmt.Sprintf("%.[2]*[1]f %[3]v", order.Size, prec.Size, asset),
					fmt.Sprintf("%[3]v %.[2]*[1]f", (order.Price * order.Size), prec.Price, quote),
				})
			}

			fmt.Println(writer.Render())
		}

		return nil
	},
}
