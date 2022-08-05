package internal

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/svanas/ladder/exchange"
	"github.com/svanas/ladder/precision"
)

// given a number of input `steps`, this function will calculate how much of the asset we will sell
func Simulate(start_with_size, mult float64, steps int) float64 {
	// this is the very 1st step we will always make
	result := start_with_size
	// go over every other step
	for step := 1; step < steps; step++ {
		result += start_with_size * (1 + (float64(step) * (mult - 1)))
	}
	return result
}

// compute every order
func Orders(start_at_price, stop_at_price, start_with_size, mult, size float64, steps int, prec *exchange.Precision) (result []Order) {
	var cumulative_size float64

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

		result = append(result, Order{
			Price: precision.Round(current_price, prec.Price),
			Size:  precision.Round(current_size, prec.Size),
		})

		current_size = start_with_size * (1 + (float64(step+1) * (mult - 1)))
		current_price += delta
	}

	return result
}

// print every order to standard output
func Print(asset, quote string, start_at_price, stop_at_price, start_with_size, mult, size float64, steps int, prec *exchange.Precision) {
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

		current_size = start_with_size * (1 + (float64(step+1) * (mult - 1)))
		current_price += delta
	}

	tbl.AppendSeparator()
	tbl.AppendRow(table.Row{"TOTAL", "",
		fmt.Sprintf("%.[2]*[1]f", cumulative_size, prec.Size),
		fmt.Sprintf("%[3]v %.[2]*[1]f", cumulative_value, prec.Price, quote),
	})

	fmt.Println(tbl.Render())
}
