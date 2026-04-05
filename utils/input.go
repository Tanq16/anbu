package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var stdinScanner *bufio.Scanner

type singleLineModel struct {
	textInput textinput.Model
	header    string
	width     int
	quitting  bool
	output    string
}

func newSingleLineModel(header string, placeholder string, masked bool) singleLineModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	if masked {
		ti.EchoMode = textinput.EchoPassword
	}
	ti.Focus()
	ti.Prompt = " > "
	return singleLineModel{
		textInput: ti,
		header:    header,
	}
}

func (m singleLineModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m singleLineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter":
			m.output = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m singleLineModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	wrapper := lipgloss.NewStyle().Width(m.width - 2)
	var view strings.Builder
	if m.header != "" {
		view.WriteString(wrapper.Render(m.header))
		view.WriteString("\n")
	}
	view.WriteString(m.textInput.View())
	return tea.NewView(view.String())
}

type multiLineModel struct {
	textInput textarea.Model
	header    string
	width     int
	quitting  bool
	output    string
}

func newMultiLineModel(header string, placeholder string) multiLineModel {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.MaxHeight = 0
	ta.Focus()
	styles := ta.Styles()
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(styles)
	ta.SetHeight(12)
	ta.ShowLineNumbers = true
	ta.Prompt = " ┃ "
	return multiLineModel{
		textInput: ta,
		header:    header,
	}
}

func (m multiLineModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m multiLineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.output = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		case "alt+enter":
			m.output = strings.TrimSpace(m.textInput.Value())
			m.quitting = true
			return m, tea.Quit
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m multiLineModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	wrapper := lipgloss.NewStyle().Width(m.width - 2)
	var view strings.Builder
	if m.header != "" {
		headerText := m.header + FDebug(" (Press Ctrl+D to submit)")
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
	p := tea.NewProgram(newSingleLineModel(prompt, placeholder, false))
	m, err := p.Run()
	if err != nil {
		PrintError("Input error", err)
		return ""
	}
	if model, ok := m.(singleLineModel); ok {
		return model.output
	}
	return ""
}

func PromptPassword(prompt string) string {
	if GlobalForAIFlag {
		result, err := ReadPipedLine()
		if err != nil {
			PrintError("Piped input error", err)
			return ""
		}
		return result
	}
	LineBreak()
	p := tea.NewProgram(newSingleLineModel(prompt, "", true))
	m, err := p.Run()
	if err != nil {
		PrintError("Input error", err)
		return ""
	}
	if model, ok := m.(singleLineModel); ok {
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
	p := tea.NewProgram(newMultiLineModel(prompt, placeholder))
	m, err := p.Run()
	if err != nil {
		PrintError("Input error", err)
		return ""
	}
	if model, ok := m.(multiLineModel); ok {
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
	p := tea.NewProgram(newSingleLineModel(sb.String(), "Paste URL here", false))
	m, err := p.Run()
	if err != nil {
		PrintError("Bubbletea error", err)
		return ""
	}
	if finalModel, ok := m.(singleLineModel); ok {
		return finalModel.output
	}
	return ""
}
