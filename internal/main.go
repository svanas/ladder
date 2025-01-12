package internal

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/svanas/ladder/exchange"
	"github.com/svanas/ladder/precision"
)

// given a number of input `steps`, this function will calculate how much of the BASE asset we will sell
func SimulateSell(start_with_size, mult float64, steps int) float64 {
	// this is the very 1st step we will always make
	result := start_with_size
	// go over every other step
	for step := 1; step < steps; step++ {
		result += start_with_size * (1 + (float64(step) * (mult - 1)))
	}
	return result
}

// given a number of input `steps`, this function will calculate how much of the QUOTE asset we will buy
func SimulateBuy(start_at_price, stop_at_price, start_with_size, mult float64, steps int) float64 {
	result := 0.0
	// this is the very 1st order we will always make
	current_price := start_at_price
	current_size := start_with_size / start_at_price
	// calculate the price delta between the orders
	delta := (stop_at_price - start_at_price) / (float64(steps) - 1)
	// calculate how much of the QUOTE asset we will buy
	for step := 0; step < steps; step++ {
		result += (current_size * current_price)
		current_size = (start_with_size / start_at_price) * (1 + (float64(step+1) * (mult - 1)))
		current_price += delta
	}
	return result
}

// compute every order
func Orders(start_at_price, stop_at_price, start_with_size, mult, size float64, steps int, prec exchange.Precision) (result []exchange.Order) {
	var cumulative_size float64 = 0

	// this is the very 1st order we will always make
	current_price := start_at_price
	current_size := start_with_size

	// calculate the price delta between the orders
	delta := (stop_at_price - start_at_price) / (float64(steps) - 1)

	for step := 0; step < steps; step++ {
		// sweeping the dust from your wallet
		if step == (steps - 1) {
			current_size = size - cumulative_size
		}
		cumulative_size += current_size

		result = append(result, exchange.Order{
			Price: precision.Round(current_price, prec.Price),
			Size:  precision.Round(current_size, prec.Size),
		})

		current_size = start_with_size * (1 + (float64(step+1) * (mult - 1)))
		current_price += delta
	}

	return result
}

// print every order to standard output
func Print(asset, quote string, start_at_price, stop_at_price, start_with_size, mult, size float64, steps int, prec exchange.Precision) {
	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{"", "Price", "Size", "Value"})

	var (
		cumulative_size  float64 = 0
		cumulative_value float64 = 0
	)

	// this is the very 1st order we will always make
	current_price := start_at_price
	current_size := start_with_size

	// calculate the price delta between the orders
	delta := (stop_at_price - start_at_price) / (float64(steps) - 1)

	for step := 0; step < steps; step++ {
		// sweeping the dust from your wallet
		if step == (steps - 1) {
			current_size = size - cumulative_size
		}

		cumulative_size += current_size
		cumulative_value += current_price * current_size

		tbl.AppendRow(table.Row{step + 1,
			fmt.Sprintf("%[3]v %.[2]*[1]f", current_price, prec.Price, quote),
			fmt.Sprintf("%.[2]*[1]f %[3]v", current_size, prec.Size, asset),
			fmt.Sprintf("%[3]v %.[2]*[1]f", (current_price * current_size), prec.Price, quote),
		})

		current_size = start_with_size * (1 + (float64(step+1) * (mult - 1)))
		current_price += delta
	}

	tbl.AppendSeparator()
	tbl.AppendRow(table.Row{"TOTAL", "",
		fmt.Sprintf("%.[2]*[1]f %[3]v", cumulative_size, prec.Size, asset),
		fmt.Sprintf("%[3]v %.[2]*[1]f", cumulative_value, prec.Price, quote),
	})

	fmt.Println(tbl.Render())
}
