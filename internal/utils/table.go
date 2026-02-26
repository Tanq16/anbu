package utils

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Padding(0, 1)
	cellStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 1)
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
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
	t.table = table.New().
		Headers(headers...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
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
	PrintGeneric(t.FormatTable(useMarkdown))
}
