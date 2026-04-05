package utils

import "charm.land/lipgloss/v2"

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(9))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12))
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))
	streamStyle  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))
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
