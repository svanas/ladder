//lint:file-ignore ST1006 receiver name should be a reflection of its identity; don't use generic names such as "this" or "self"
package internal

import (
	"fmt"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/svanas/ladder/answer"
)

type Order struct {
	Price float64
	Size  float64
}

func (self *Order) Prompt(market string) answer.Answer {
	const TITLE = "Open this order?"

	tbl := table.NewWriter()
	tbl.AppendHeader(table.Row{TITLE, TITLE}, table.RowConfig{AutoMerge: true})
	tbl.AppendRows([]table.Row{
		{"Market", market},
		{"Price", strconv.FormatFloat(self.Price, 'f', -1, 64)},
		{"Size", strconv.FormatFloat(self.Size, 'f', -1, 64)},
	})
	fmt.Println(tbl.Render())

	return answer.Ask()
}
