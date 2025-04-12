package utils

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mattn/go-isatty"
)

var boxChars = map[string]string{
	"topLeft":     "┌",
	"topRight":    "┐",
	"bottomLeft":  "└",
	"bottomRight": "┘",
	"horizontal":  "─",
	"vertical":    "│",
	"leftT":       "├",
	"rightT":      "┤",
	"topT":        "┬",
	"bottomT":     "┴",
	"cross":       "┼",
}

type MarkdownTable struct {
	Caption string
	Headers []string
	Rows    [][]string
}

// OutMDFile writes a markdown table to a file
func (table MarkdownTable) OutMDFile(outputPath string) error {
	formatted, err := formatMDTable(table)
	if err != nil {
		return fmt.Errorf("error formatting table: %w", err)
	}
	var outFile *os.File
	_, err = os.Stat(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the file if it doesn't exist
			outFile, err = os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
			defer outFile.Close()
		} else {
			return fmt.Errorf("error checking file: %w", err)
		}
	} else {
		// If the file exists, open to rewrite
		outFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer outFile.Close()
	}
	// Write the formatted table to the file
	_, err = outFile.WriteString(formatted)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}

// OutMDPrint prints a markdown table to console
func (table MarkdownTable) OutMDPrint(innerDivide bool) error {
	formatted, err := formatTerminalTable(table, innerDivide)
	if err != nil {
		return fmt.Errorf("error formatting table: %w", err)
	}
	fmt.Print(formatted)
	return nil
}

func getTerminalWidth() int {
	if widthStr := os.Getenv("COLUMNS"); widthStr != "" {
		if width, err := strconv.Atoi(widthStr); err == nil && width > 0 {
			return width
		}
	}
	// using syscall on unix systems
	if runtime.GOOS != "windows" {
		type windowSize struct {
			rows    uint16
			cols    uint16
			xpixels uint16
			ypixels uint16
		}
		ws := &windowSize{}
		if isTerm := isatty.IsTerminal(os.Stdout.Fd()); isTerm {
			if _, _, err := syscall.Syscall(syscall.SYS_IOCTL,
				os.Stdout.Fd(),
				uintptr(syscall.TIOCGWINSZ),
				uintptr(unsafe.Pointer(ws))); err == 0 {
				return int(ws.cols)
			}
		}
	}
	// Fallback to a default
	return 120
}

func wrapText(text string, width int) []string {
	if width <= 0 || len(text) <= width {
		return []string{text}
	}
	var lines []string
	var currentLine string
	words := strings.Fields(text)
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			if len(word) > width {
				for len(word) > 0 {
					if len(word) <= width {
						currentLine = word
						break
					}
					lines = append(lines, word[:width])
					word = word[width:]
				}
			} else {
				currentLine = word
			}
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func formatTerminalTable(table MarkdownTable, innerDivide bool) (string, error) {
	var output strings.Builder
	termWidth := getTerminalWidth()
	if table.Caption != "" {
		output.WriteString(table.Caption + "\n\n")
	}
	// Dynamically determine wrapped width
	colWidths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	borderChars := 1 + len(colWidths)
	paddingChars := len(colWidths) * 2
	contentWidth := 0
	for _, w := range colWidths {
		contentWidth += w
	}
	totalWidth := borderChars + paddingChars + contentWidth
	// adjust column widths proportionally
	if totalWidth > termWidth && termWidth > (borderChars+paddingChars+len(colWidths)) {
		availableWidth := termWidth - borderChars - paddingChars
		totalContentWidth := contentWidth
		for i := range colWidths {
			ratio := float64(colWidths[i]) / float64(totalContentWidth)
			colWidths[i] = int(math.Floor(ratio * float64(availableWidth)))
		}
	}
	// Recalculate total width after adjustments
	totalWidth = 1 // for left border
	for _, width := range colWidths {
		totalWidth += width + 2 // +2 for padding
	}
	totalWidth += len(colWidths) - 1

	// Start printing the table

	output.WriteString(boxChars["topLeft"])
	for i, width := range colWidths {
		output.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			output.WriteString(boxChars["topT"])
		}
	}
	output.WriteString(boxChars["topRight"] + "\n")
	wrappedHeaders := make([][]string, len(table.Headers))
	maxHeaderLines := 1
	// Wrap headers
	for i, header := range table.Headers {
		wrappedHeaders[i] = wrapText(header, colWidths[i])
		if len(wrappedHeaders[i]) > maxHeaderLines {
			maxHeaderLines = len(wrappedHeaders[i])
		}
	}
	// Print headers with wrapping
	for line := range maxHeaderLines {
		output.WriteString(boxChars["vertical"])
		for i, wrappedHeader := range wrappedHeaders {
			headerLine := ""
			if line < len(wrappedHeader) {
				headerLine = wrappedHeader[line]
			}
			format := fmt.Sprintf(" %%-%ds ", colWidths[i])
			paddedHeader := fmt.Sprintf(format, headerLine)
			paddedHeader = OutColors["bold"] + OutColors["italic"] + paddedHeader + OutColors["reset"]
			output.WriteString(paddedHeader)
			if i < len(wrappedHeaders)-1 {
				output.WriteString(boxChars["vertical"])
			}
		}
		output.WriteString(boxChars["vertical"] + "\n")
	}
	// Header divider
	output.WriteString(boxChars["leftT"])
	for i, width := range colWidths {
		output.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			output.WriteString(boxChars["cross"])
		}
	}
	output.WriteString(boxChars["rightT"] + "\n")

	// Print data rows
	for r, row := range table.Rows {
		wrappedCells := make([][]string, len(row))
		maxLines := 1
		// Wrap cells
		for i, cell := range row {
			if i < len(colWidths) {
				wrappedCells[i] = wrapText(cell, colWidths[i])
				if len(wrappedCells[i]) > maxLines {
					maxLines = len(wrappedCells[i])
				}
			}
		}
		// Print rows with wrapping
		for line := range maxLines {
			output.WriteString(boxChars["vertical"])
			for i := range row {
				if i < len(colWidths) {
					cellLine := ""
					if line < len(wrappedCells[i]) {
						cellLine = wrappedCells[i][line]
					}
					format := fmt.Sprintf(" %%-%ds ", colWidths[i])
					output.WriteString(fmt.Sprintf(format, cellLine))
					if i < len(row)-1 && i < len(colWidths)-1 {
						output.WriteString(boxChars["vertical"])
					}
				}
			}
			output.WriteString(boxChars["vertical"] + "\n")
		}
		// Print row divider
		if r < len(table.Rows)-1 && innerDivide {
			output.WriteString(boxChars["leftT"])
			for i, width := range colWidths {
				output.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
				if i < len(colWidths)-1 {
					output.WriteString(boxChars["cross"])
				}
			}
			output.WriteString(boxChars["rightT"] + "\n")
		}
	}

	// Print bottom border
	output.WriteString(boxChars["bottomLeft"])
	for i, width := range colWidths {
		output.WriteString(strings.Repeat(boxChars["horizontal"], width+2))
		if i < len(colWidths)-1 {
			output.WriteString(boxChars["bottomT"])
		}
	}
	output.WriteString(boxChars["bottomRight"] + "\n")
	return output.String(), nil
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
