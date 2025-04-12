package utils

import (
	"fmt"
	"os"
	"strings"
)

var OutColors = map[string]string{
	"red":            "\033[31m",
	"green":          "\033[32m",
	"yellow":         "\033[33m",
	"blue":           "\033[34m",
	"purple":         "\033[35m",
	"cyan":           "\033[36m",
	"white":          "\033[37m",
	"black":          "\033[30m",
	"grey":           "\033[90m",
	"brightRed":      "\033[91m",
	"brightGreen":    "\033[92m",
	"brightYellow":   "\033[93m",
	"brightBlue":     "\033[94m",
	"brightPurple":   "\033[95m",
	"brightCyan":     "\033[96m",
	"brightWhite":    "\033[97m",
	"bgRed":          "\033[41m",
	"bgGreen":        "\033[42m",
	"bgYellow":       "\033[43m",
	"bgBlue":         "\033[44m",
	"bgPurple":       "\033[45m",
	"bgCyan":         "\033[46m",
	"bgWhite":        "\033[47m",
	"bgBlack":        "\033[40m",
	"bgGrey":         "\033[100m",
	"bgBrightRed":    "\033[101m",
	"bgBrightGreen":  "\033[102m",
	"bgBrightYellow": "\033[103m",
	"bgBrightBlue":   "\033[104m",
	"bgBrightPurple": "\033[105m",
	"bgBrightCyan":   "\033[106m",
	"bgBrightWhite":  "\033[107m",
	"bold":           "\033[1m",
	"dim":            "\033[2m",
	"italic":         "\033[3m",
	"underline":      "\033[4m",
	"blink":          "\033[5m",
	"reset":          "\033[0m",
}

func OutClearLines(n int) {
	if n == 0 {
		fmt.Print("\033[H\033[2J") // Clear the screen
	}
	fmt.Printf("\033[%dA", n)
}

func OutSuccess(msg string) {
	fmt.Printf("%s%s%s\n", OutColors["blue"], msg, OutColors["reset"])
}

func OutError(msg string) {
	fmt.Printf("%s%s%s\n", OutColors["red"], msg, OutColors["reset"])
}

func OutWarning(msg string) {
	fmt.Printf("%s%s%s\n", OutColors["yellow"], msg, OutColors["reset"])
}

func OutInfo(msg string) {
	fmt.Printf("%s%s%s\n", OutColors["green"], msg, OutColors["reset"])
}

func OutDebug(msg string) {
	fmt.Printf("%s%s%s\n", OutColors["grey"], msg, OutColors["reset"])
}

type MarkdownTable struct {
	Caption string
	Headers []string
	Rows    [][]string
}

// OutMDFile writes a markdown table to a file
func (table MarkdownTable) OutMDFile(outputFile *os.File) error {
	formatted, err := formatMDTable(table)
	if err != nil {
		return fmt.Errorf("error formatting table: %w", err)
	}
	_, err = outputFile.WriteString(formatted)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}

// OutMDPrint prints a markdown table to console
func (table MarkdownTable) OutMDPrint() error {
	formatted, err := formatMDTable(table)
	if err != nil {
		return fmt.Errorf("error formatting table: %w", err)
	}
	fmt.Print(formatted)
	return nil
}

func formatMDTable(table MarkdownTable) (string, error) {
	var output strings.Builder
	if table.Caption != "" {
		output.WriteString("# " + table.Caption + "\n\n")
	}
	// Dynamically determine column width
	colWidths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	// Write header
	headerRow := "|"
	dividerRow := "|"
	for i, header := range table.Headers {
		formatter := fmt.Sprintf(" %%-%ds |", colWidths[i])
		headerRow += fmt.Sprintf(formatter, header)
		dividerRow += fmt.Sprintf(" %s |", strings.Repeat("-", colWidths[i]))
	}
	headerRow += "\n"
	dividerRow += "\n"
	output.WriteString(headerRow)
	output.WriteString(dividerRow)
	// Write rows
	for _, row := range table.Rows {
		rowText := "|"
		for i, cell := range row {
			formatter := fmt.Sprintf(" %%-%ds |", colWidths[i])
			rowText += fmt.Sprintf(formatter, cell)
		}
		rowText += "\n"
		output.WriteString(rowText)
	}
	output.WriteString("\n\n")
	return output.String(), nil
}
