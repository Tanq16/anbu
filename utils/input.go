package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var ErrNoTerminal = errors.New("no interactive terminal")

const stdinAnnotation = "stdin"
const stdinResolvedAnnotation = "stdin-resolved"

func MarkStdinLine(cmd *cobra.Command, name string) error {
	return cmd.Flags().SetAnnotation(name, stdinAnnotation, []string{"line"})
}

func MarkStdinStream(cmd *cobra.Command, name string) error {
	return cmd.Flags().SetAnnotation(name, stdinAnnotation, []string{"stream"})
}

func ResolveStdin(cmd *cobra.Command) error {
	var target *pflag.Flag
	var mode string
	var err error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		modes, ok := f.Annotations[stdinAnnotation]
		if !ok || len(modes) == 0 || !f.Changed || f.Value.String() != "-" {
			return
		}
		if target != nil {
			err = fmt.Errorf("only one flag can read stdin: --%s and --%s were both given -", target.Name, f.Name)
			return
		}
		target, mode = f, modes[0]
	})
	if err != nil || target == nil {
		return err
	}
	if StdinIsTerminal {
		return fmt.Errorf("--%s was given - but nothing is piped into stdin", target.Name)
	}

	var value string
	if mode == "line" {
		line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}
		value = strings.TrimRight(line, "\r\n")
	} else {
		data, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			return readErr
		}
		value = strings.TrimRight(string(data), "\r\n")
	}
	if value == "" {
		return fmt.Errorf("--%s was given - but stdin was empty", target.Name)
	}
	if err := target.Value.Set(value); err != nil {
		return err
	}
	return cmd.Flags().SetAnnotation(target.Name, stdinResolvedAnnotation, []string{"true"})
}

func ReadFileFlag(cmd *cobra.Command, name string) (string, error) {
	f := cmd.Flags().Lookup(name)
	if f == nil {
		return "", nil
	}
	value := f.Value.String()
	if value == "" {
		return "", nil
	}
	if resolved, ok := f.Annotations[stdinResolvedAnnotation]; ok && len(resolved) > 0 {
		return value, nil
	}
	data, err := os.ReadFile(value)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

type inputModel struct {
	textInput textinput.Model
	done      bool
	value     string
	initCmd   tea.Cmd
}

func (m inputModel) Init() tea.Cmd {
	return m.initCmd
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			m.value = m.textInput.Value()
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(m.textInput.View())
}

func PromptInput(prompt string, placeholder string) (string, error) {
	if !StdinIsTerminal {
		return "", ErrNoTerminal
	}

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = prompt + " "
	m := inputModel{textInput: ti, initCmd: ti.Focus()}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(finalModel.(inputModel).value), nil
}

func PromptPassword(prompt string) (string, error) {
	if !StdinIsTerminal {
		return "", ErrNoTerminal
	}

	ti := textinput.New()
	ti.Placeholder = "••••••••"
	ti.Prompt = prompt + " "
	ti.EchoMode = textinput.EchoPassword
	m := inputModel{textInput: ti, initCmd: ti.Focus()}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	return finalModel.(inputModel).value, nil
}

type textAreaModel struct {
	textarea textarea.Model
	done     bool
	value    string
	initCmd  tea.Cmd
}

func (m textAreaModel) Init() tea.Cmd {
	return m.initCmd
}

func (m textAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+d":
			m.value = m.textarea.Value()
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m textAreaModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	return tea.NewView(m.textarea.View() + "\n(Ctrl+D to submit, Esc to cancel)")
}

func PromptTextArea(prompt string, placeholder string) (string, error) {
	if !StdinIsTerminal {
		return "", ErrNoTerminal
	}

	PrintInfo(prompt)

	ta := textarea.New()
	ta.Placeholder = placeholder
	m := textAreaModel{textarea: ta, initCmd: ta.Focus()}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(finalModel.(textAreaModel).value), nil
}

var (
	selectCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12)).Bold(true)
	selectSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10))
)

type selectModel struct {
	label   string
	options []string
	cursor  int
	chosen  int
	done    bool
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = m.cursor
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.chosen = -1
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	var b strings.Builder
	b.WriteString(m.label + "\n")
	for i, opt := range m.options {
		if i == m.cursor {
			b.WriteString(selectCursorStyle.Render("> "+opt) + "\n")
		} else {
			b.WriteString("  " + opt + "\n")
		}
	}
	return tea.NewView(b.String())
}

func PromptSelect(label string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, nil
	}
	if !StdinIsTerminal {
		return -1, ErrNoTerminal
	}

	m := selectModel{label: label, options: options, chosen: -1}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return -1, err
	}
	return finalModel.(selectModel).chosen, nil
}

type multiSelectModel struct {
	label     string
	options   []string
	cursor    int
	selected  map[int]bool
	cancelled bool
	done      bool
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.cancelled = true
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}
	var b strings.Builder
	b.WriteString(m.label + " (space toggles, enter confirms)\n")
	for i, opt := range m.options {
		mark := "[ ]"
		if m.selected[i] {
			mark = selectSelectedStyle.Render("[x]")
		}
		line := mark + " " + opt
		if i == m.cursor {
			line = selectCursorStyle.Render("> ") + line
		} else {
			line = "  " + line
		}
		b.WriteString(line + "\n")
	}
	return tea.NewView(b.String())
}

func PromptMultiSelect(label string, options []string) (map[int]bool, error) {
	if len(options) == 0 {
		return nil, nil
	}
	if !StdinIsTerminal {
		return nil, ErrNoTerminal
	}

	m := multiSelectModel{label: label, options: options, selected: make(map[int]bool)}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	result := finalModel.(multiSelectModel)
	if result.cancelled {
		return nil, nil
	}
	return result.selected, nil
}

func DeviceCodeFlow(url string, userCode string) (string, error) {
	LineBreak()
	PrintInfo("Visit this URL to authorize Anbu:")
	PrintGeneric(url)
	placeholder := "Paste URL here"
	if userCode != "" {
		PrintInfo("Enter the code: " + userCode)
		PrintInfo("Press Return after you have completed the authorization in your browser")
		placeholder = ""
	} else {
		PrintInfo("After authorizing, you will be redirected to a 'localhost' URL.")
		PrintInfo("Copy the *entire* URL from your browser and paste it below:")
	}
	value, err := PromptInput("", placeholder)
	if err != nil {
		return "", err
	}
	return value, nil
}
