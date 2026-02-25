package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Table struct {
	Headers []string
	Rows    [][]string
	table   *table.Table
}

func NewTable(headers []string) *Table {
	t := &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
	t.table = table.New().Headers(headers...)
	t.table = t.table.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Padding(0, 1)
		}
		return lipgloss.NewStyle().Padding(0, 1)
	})
	return t
}

func (t *Table) reconcileRows() {
	if len(t.Rows) == 0 {
		return
	}
	for _, row := range t.Rows {
		t.table.Row(row...)
	}
}

func (t *Table) FormatTable(useMarkdown bool) string {
	t.reconcileRows()
	if useMarkdown {
		return t.table.Border(lipgloss.MarkdownBorder()).String()
	}
	return t.table.String()
}

func (t *Table) PrintTable(useMarkdown bool) {
	fmt.Println(t.FormatTable(useMarkdown))
}
