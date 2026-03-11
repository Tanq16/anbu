package utils

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
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

func (t *Table) formatMarkdown() string {
	if len(t.Headers) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("| " + strings.Join(escapeCells(t.Headers), " | ") + " |")
	sb.WriteByte('\n')
	seps := make([]string, len(t.Headers))
	for i := range seps {
		seps[i] = "---"
	}
	sb.WriteString("| " + strings.Join(seps, " | ") + " |")
	for _, row := range t.Rows {
		sb.WriteByte('\n')
		sb.WriteString("| " + strings.Join(escapeCells(row), " | ") + " |")
	}
	return sb.String()
}

func escapeCells(cells []string) []string {
	escaped := make([]string, len(cells))
	for i, cell := range cells {
		cell = strings.ReplaceAll(cell, "|", "\\|")
		cell = strings.ReplaceAll(cell, "\n", " ")
		escaped[i] = cell
	}
	return escaped
}

func (t *Table) FormatTable(useMarkdown bool) string {
	if useMarkdown {
		return t.formatMarkdown()
	}
	t.reconcileRows()
	return t.table.String()
}

func (t *Table) PrintTable(useMarkdown bool) {
	if GlobalForAIFlag {
		useMarkdown = true
	}
	PrintGeneric(t.FormatTable(useMarkdown))
}
