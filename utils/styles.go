package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

var GlobalDebugFlag bool

var (
	// Core styles
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))  // dark green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // yellow
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // blue
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))  // cyan
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // light grey
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // grey

	// Additional config
	basePadding = 2
)

var StyleSymbols = map[string]string{
	"pass":    "✓",
	"fail":    "✗",
	"warning": "!",
	"pending": "◉",
	"info":    "ℹ",
	"arrow":   "→",
	"bullet":  "•",
	"dot":     "·",
	"hline":   "━",
}

func PrintSuccess(text string) {
	if !GlobalDebugFlag {
		fmt.Println(successStyle.Render(text))
	} else {
		log.Info().Msg(text)
	}
}
func PrintError(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(errorStyle.Render(text))
	} else {
		log.Error().Err(err).Msg(text)
	}
}
func PrintFatal(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(errorStyle.Render(text))
		os.Exit(1)
	} else {
		log.Fatal().Err(err).Msg(text)
	}
}
func PrintWarning(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(warningStyle.Render(text))
	} else {
		log.Warn().Err(err).Msg(text)
	}
}
func PrintInfo(text string) {
	if !GlobalDebugFlag {
		fmt.Println(infoStyle.Render(text))
	} else {
		log.Info().Msg(text)
	}
}
func PrintDebug(text string) {
	if !GlobalDebugFlag {
		fmt.Println(debugStyle.Render(text))
	} else {
		log.Debug().Msg(text)
	}
}
func PrintStream(text string) {
	if !GlobalDebugFlag {
		fmt.Println(streamStyle.Render(text))
	} else {
		log.Debug().Msg(text)
	}
}
func PrintGeneric(text string) {
	if !GlobalDebugFlag {
		fmt.Println(text)
	} else {
		log.Debug().Msg(text)
	}
}
func FSuccess(text string) string {
	if !GlobalDebugFlag {
		return successStyle.Render(text)
	} else {
		return text
	}
}
func FError(text string) string {
	if !GlobalDebugFlag {
		return errorStyle.Render(text)
	} else {
		return text
	}
}
func FWarning(text string) string {
	if !GlobalDebugFlag {
		return warningStyle.Render(text)
	} else {
		return text
	}
}
func FInfo(text string) string {
	if !GlobalDebugFlag {
		return infoStyle.Render(text)
	} else {
		return text
	}
}
func FDebug(text string) string {
	if !GlobalDebugFlag {
		return debugStyle.Render(text)
	} else {
		return text
	}
}
func FStream(text string) string {
	if !GlobalDebugFlag {
		return streamStyle.Render(text)
	} else {
		return text
	}
}
func FGeneric(text string) string {
	if !GlobalDebugFlag {
		return text
	} else {
		return text
	}
}

func LineBreak() {
	fmt.Println()
}
func ClearTerminal(lines int) {
	if lines > 0 {
		fmt.Printf("\033[%dA\r\033[K", lines)
	}
}

func readSingleLineInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(text), nil
}

func readMultilineInput() (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOF" {
			break
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	text := sb.String()
	if len(text) > 0 {
		text = text[:len(text)-1]
	}
	return text, nil
}

func InputWithClear(prompt string) string {
	fmt.Print("\0337")
	fmt.Print(prompt)
	input, err := readSingleLineInput()
	if err != nil {
		PrintError("error reading input", err)
		return ""
	}
	fmt.Print("\0338")
	fmt.Print("\0338\033[J")
	return input
}

func MultilineInputWithClear(prompt string) string {
	fmt.Print("\0337")
	fmt.Println(prompt)
	fmt.Print("Type 'EOF' on a new line to finish:")
	input, err := readMultilineInput()
	if err != nil {
		PrintError("error reading input", err)
		return ""
	}
	fmt.Print("\0338")
	fmt.Print("\0338\033[J")
	return input
}

func DeviceCodeFlow(url string, userCode string) string {
	fmt.Print("\0337")
	PrintDebug("Visit this URL to authorize Anbu:\n\n")
	PrintInfo(url + "\n\n")
	if userCode != "" {
		PrintDebug("Enter the code: " + userCode + "\n\n")
		PrintDebug("Press Return after you have completed the authorization in your browser")
	} else {
		PrintDebug("After authorizing, you will be redirected to a 'localhost' URL.\n\n")
		PrintDebug("Copy the *entire* URL from your browser and paste it here:")
	}
	input, err := readSingleLineInput()
	if err != nil {
		PrintError("error reading input", err)
		return ""
	}
	input = strings.TrimSpace(input)
	fmt.Print("\0338")
	fmt.Print("\0338\033[J")
	return input
}
