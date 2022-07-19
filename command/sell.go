package command

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/svanas/ladder/flag"
)

func init() {
	sellCommand.Flags().String(FLAG_ASSET, "BTC", "name of the asset you will want to sell")
	sellCommand.Flags().String(FLAG_QUOTE, "USDT", "name of the asset you will want to receive")

	sellCommand.Flags().Float64(START_AT_PRICE, 0, "price where you will want to start selling at")
	sellCommand.Flags().Float64(STOP_AT_PRICE, 0, "price where you will want to stop selling")
	sellCommand.Flags().Float64(START_WITH_SIZE, 0, "size of your first sell order")

	sellCommand.Flags().Float64(FLAG_MULT, 1.025, "multiplier that defines the distance between your orders")
	sellCommand.Flags().Float64(FLAG_SIZE, 0, "the quantity you will want to sell")

	rootCommand.AddCommand(sellCommand)
}

// given a number of input `steps`, this function will calculate how much of the asset we will sell
func simulate(start_with_size, mult float64, steps int) float64 {
	size := start_with_size
	// this is the very 1st step we will always make
	result := size
	// go over every other step
	for step := 1; step < steps; step++ {
		size = size * mult
		result += size
	}
	return result
}

type Prec struct {
	Price int
	Size  int
}

// print every order to standard output
func print(asset, quote string, start_at_price, stop_at_price, start_with_size, mult, size float64, steps int, prec *Prec) {
	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{"", "Price", "Size", "Value"})

	var (
		cumulative_size  float64
		cumulative_value float64
	)

	// this is the very 1st order we will always make
	current_price := start_at_price
	current_size := start_with_size

	// calculate the price delta between the orders
	delta := (stop_at_price - start_at_price) / (float64(steps) - 1)

	for step := 0; step < steps; step++ {
		if step == (steps - 1) {
			current_size = size - cumulative_size
		}

		cumulative_size += current_size
		cumulative_value += current_price * current_size

		tbl.AppendRow(table.Row{step + 1,
			fmt.Sprintf("%[3]v %.[2]*[1]f", current_price, prec.Price, asset),
			fmt.Sprintf("%.[2]*[1]f", current_size, prec.Size),
			fmt.Sprintf("%[3]v %.[2]*[1]f", (current_price * current_size), prec.Price, quote),
		})

		current_size = current_size * mult
		current_price += delta
	}

	tbl.AppendSeparator()
	tbl.AppendRow(table.Row{"TOTAL", "",
		fmt.Sprintf("%.[2]*[1]f", cumulative_size, prec.Size),
		fmt.Sprintf("%[3]v %.[2]*[1]f", cumulative_value, prec.Price, quote),
	})

	fmt.Println(tbl.Render())
}

var sellCommand = &cobra.Command{
	Use:   "sell",
	Short: "sell your crypto asset",
	RunE: func(cmd *cobra.Command, args []string) error {
		asset, err := cmd.Flags().GetString(FLAG_ASSET)
		if err != nil {
			return err
		}

		quote, err := cmd.Flags().GetString(FLAG_QUOTE)
		if err != nil {
			return err
		}

		start_at_price, err := flag.GetFloat64(cmd, START_AT_PRICE)
		if err != nil {
			return err
		}

		stop_at_price, err := flag.GetFloat64(cmd, STOP_AT_PRICE)
		if err != nil {
			return err
		}

		start_with_size, err := flag.GetFloat64(cmd, START_WITH_SIZE)
		if err != nil {
			return err
		}

		mult, err := flag.GetFloat64(cmd, FLAG_MULT)
		if err != nil {
			return err
		}

		size, err := flag.GetFloat64(cmd, FLAG_SIZE)
		if err != nil {
			return err
		}

		steps := 2
		for simulate(start_with_size, mult, steps) < size {
			steps++
		}

		print(asset, quote, start_at_price, stop_at_price, start_with_size, mult, size, steps, &Prec{
			Price: 2,
			Size:  8,
		})

		return nil
	},
}
