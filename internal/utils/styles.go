package utils

import "github.com/charmbracelet/lipgloss"

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))  // dark green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // yellow
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))  // cyan
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // light grey
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // grey
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
