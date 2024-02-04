//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package internal

import (
	"fmt"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/svanas/ladder/answer"
	"github.com/svanas/ladder/exchange"
)

func Prompt(order *exchange.Order, market string) answer.Answer {
	const TITLE = "Open this order?"

	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{TITLE, TITLE}, table.RowConfig{AutoMerge: true})
	tbl.AppendRows([]table.Row{
		{"Market", market},
		{"Price", strconv.FormatFloat(order.Price, 'f', -1, 64)},
		{"Size", strconv.FormatFloat(order.Size, 'f', -1, 64)},
	})
	fmt.Println(tbl.Render())

	return answer.Ask()
}
