package utils

import (
	"fmt"
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

func OutSuccess(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["blue"], msg, OutColors["reset"])
}

func OutError(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["red"], msg, OutColors["reset"])
}

func OutWarning(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["yellow"], msg, OutColors["reset"])
}

func OutInfo(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["green"], msg, OutColors["reset"])
}

func OutDetail(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["purple"], msg, OutColors["reset"])
}

func OutDebug(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["grey"], msg, OutColors["reset"])
}

func OutCyan(msg string) string {
	return fmt.Sprintf("%s%s%s", OutColors["cyan"], msg, OutColors["reset"])
}
