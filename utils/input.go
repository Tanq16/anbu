package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var stdinScanner *bufio.Scanner

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
	ta.MaxHeight = 0
	ta.Focus()
	styles := ta.Styles()
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(styles)
	if multiline {
		ta.SetHeight(12)
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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+d":
			if m.multiline {
				m.output = strings.TrimSpace(m.textInput.Value())
				m.quitting = true
				return m, tea.Quit
			}
		case "enter":
			if !m.multiline {
				m.output = strings.TrimSpace(m.textInput.Value())
				m.quitting = true
				return m, tea.Quit
			}
		case "alt+enter":
			m.output = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	wrapper := lipgloss.NewStyle().Width(m.width - 2)
	var view strings.Builder
	if m.header != "" {
		headerText := m.header
		if m.multiline {
			headerText += FDebug(" (Press Ctrl+D to submit)")
		}
		view.WriteString(wrapper.Render(headerText))
		view.WriteString("\n")
	}
	view.WriteString(m.textInput.View())
	return tea.NewView(view.String())
}

func GetInput(prompt string, placeholder string) string {
	if GlobalForAIFlag {
		result, err := ReadPipedLine()
		if err != nil {
			PrintError("Piped input error", err)
			return ""
		}
		return result
	}
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
	if GlobalForAIFlag {
		result, err := ReadPipedInput()
		if err != nil {
			PrintError("Piped input error", err)
			return ""
		}
		return result
	}
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

func ReadPipedInput() (string, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read piped input: %w", err)
	}
	result := strings.TrimSpace(string(input))
	if result == "" {
		return "", fmt.Errorf("no input provided via pipe")
	}
	return result, nil
}

func ReadPipedLine() (string, error) {
	if stdinScanner == nil {
		stdinScanner = bufio.NewScanner(os.Stdin)
	}
	if stdinScanner.Scan() {
		line := strings.TrimSpace(stdinScanner.Text())
		if line == "" {
			return "", fmt.Errorf("empty line from piped input")
		}
		return line, nil
	}
	if err := stdinScanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read piped line: %w", err)
	}
	return "", fmt.Errorf("no input provided via pipe")
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
