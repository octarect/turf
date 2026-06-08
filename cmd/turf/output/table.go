package output

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = false
	t.Style().Title.Align = text.AlignCenter

	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		r := make(table.Row, len(row))
		for i, v := range row {
			r[i] = v
		}
		t.AppendRow(r)
	}

	t.Render()
	return nil
}
