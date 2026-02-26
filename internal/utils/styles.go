package utils

import "github.com/charmbracelet/lipgloss"

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
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
