package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
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
	}
	return text
}
func FError(text string) string {
	if !GlobalDebugFlag {
		return errorStyle.Render(text)
	}
	return text
}
func FWarning(text string) string {
	if !GlobalDebugFlag {
		return warningStyle.Render(text)
	}
	return text
}
func FInfo(text string) string {
	if !GlobalDebugFlag {
		return infoStyle.Render(text)
	}
	return text
}
func FDebug(text string) string {
	if !GlobalDebugFlag {
		return debugStyle.Render(text)
	}
	return text
}
func FStream(text string) string {
	if !GlobalDebugFlag {
		return streamStyle.Render(text)
	}
	return text
}
func FGeneric(text string) string {
	return text
}

func LineBreak() {
	fmt.Println()
}
func ClearTerminal(lines int) {
	if lines > 0 {
		fmt.Printf("\033[%dA\r\033[K", lines)
	}
}

type inputModel struct {
	textInput textarea.Model
	header    string
	width     int
	multiline bool
	quitting  bool
	output    string
	err       error
}

func newInputModel(header string, placeholder string, multiline bool) inputModel {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	if multiline {
		ta.SetHeight(10)
		ta.ShowLineNumbers = true
		ta.Prompt = " ┃ "
	} else {
		ta.SetHeight(1)
		ta.ShowLineNumbers = false
		ta.Prompt = " > "
	}
	return inputModel{
		textInput: ta,
		header:    header,
		multiline: multiline,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.textInput.SetWidth(msg.Width - 4)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlD:
			if m.multiline {
				m.output = strings.TrimSpace(m.textInput.Value())
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyEnter:
			if msg.Alt {
				m.output = strings.TrimSpace(m.textInput.Value())
				m.quitting = true
				return m, tea.Quit
			}
			if !m.multiline {
				m.output = strings.TrimSpace(m.textInput.Value())
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}
	wrapper := lipgloss.NewStyle().Width(m.width - 2)
	var view strings.Builder
	if m.header != "" {
		headerText := m.header
		if m.multiline {
			headerText += FStream(" (Press Ctrl+D to submit)")
		}
		view.WriteString(wrapper.Render(headerText))
		view.WriteString("\n")
	}
	view.WriteString(m.textInput.View())
	return view.String()
}

func GetInput(prompt string, placeholder string) string {
	LineBreak()
	p := tea.NewProgram(newInputModel(prompt, placeholder, false))
	m, err := p.Run()
	if err != nil {
		PrintError("Input error", err)
		return ""
	}
	if model, ok := m.(inputModel); ok {
		return model.output
	}
	return ""
}

func GetMultilineInput(prompt string, placeholder string) string {
	LineBreak()
	p := tea.NewProgram(newInputModel(prompt, placeholder, true))
	m, err := p.Run()
	if err != nil {
		PrintError("Input error", err)
		return ""
	}
	if model, ok := m.(inputModel); ok {
		return model.output
	}
	return ""
}

func DeviceCodeFlow(url string, userCode string) string {
	LineBreak()
	var sb strings.Builder
	sb.WriteString(FDebug("Visit this URL to authorize Anbu:") + "\n")
	sb.WriteString(FGeneric(url) + "\n\n")
	if userCode != "" {
		sb.WriteString(FDebug("Enter the code: "+userCode) + "\n")
		sb.WriteString(FDebug("Press Return after you have completed the authorization in your browser"))
	} else {
		sb.WriteString(FDebug("After authorizing, you will be redirected to a 'localhost' URL.") + "\n")
		sb.WriteString(FDebug("Copy the *entire* URL from your browser and paste it below:"))
	}
	p := tea.NewProgram(newInputModel(sb.String(), "Paste URL here", false))
	m, err := p.Run()
	if err != nil {
		PrintError("Bubbletea error", err)
		return ""
	}
	if finalModel, ok := m.(inputModel); ok {
		return finalModel.output
	}
	return ""
}
