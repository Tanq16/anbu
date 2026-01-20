package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
		fmt.Println(text)
	}
}
func PrintError(text string) {
	if !GlobalDebugFlag {
		fmt.Println(errorStyle.Render(text))
	} else {
		fmt.Println(text)
	}
}
func PrintWarning(text string) {
	if !GlobalDebugFlag {
		fmt.Println(warningStyle.Render(text))
	} else {
		fmt.Println(text)
	}
}
func PrintInfo(text string) {
	if !GlobalDebugFlag {
		fmt.Println(infoStyle.Render(text))
	} else {
		fmt.Println(text)
	}
}
func PrintDebug(text string) {
	if !GlobalDebugFlag {
		fmt.Println(debugStyle.Render(text))
	} else {
		fmt.Println(text)
	}
}
func PrintStream(text string) {
	if !GlobalDebugFlag {
		fmt.Println(streamStyle.Render(text))
	} else {
		fmt.Println(text)
	}
}
func PrintGeneric(text string) {
	if !GlobalDebugFlag {
		fmt.Println(text)
	} else {
		fmt.Println(text)
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
