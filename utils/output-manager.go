package utils

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type Manager struct {
	program *tea.Program
	done    chan struct{}
}

type progressMsg struct {
	current int64
	total   int64
	message string
}

type progressModel struct {
	current  int64
	total    int64
	message  string
	bar      progress.Model
	quitting bool
}

func initialModel() progressModel {
	bar := progress.New(progress.WithFillCharacters('━', ' '))
	bar.FullColor = "250" // debug style color
	return progressModel{
		current: 0,
		total:   100,
		message: "Starting progress bar",
		bar:     bar,
	}
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.bar.Width = min(msg.Width-10, 80)
		return m, nil
	case progressMsg:
		m.current = msg.current
		m.total = msg.total
		m.message = msg.message
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m progressModel) View() string {
	if m.quitting {
		return ""
	}
	if m.total == 0 {
		m.total = 1
	}
	percent := float64(m.current) / float64(m.total)
	return fmt.Sprintf("\n%s\n%s\n",
		debugStyle.Render(m.message),
		m.bar.ViewAs(percent),
	)
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddProgressBarToStream(outof, final int64, text string) {
	if m.program != nil {
		m.program.Send(progressMsg{
			current: outof,
			total:   final,
			message: text,
		})
	}
}

func (m *Manager) StartDisplay() {
	m.program = tea.NewProgram(initialModel())
	m.done = make(chan struct{})
	go func() {
		m.program.Run()
		close(m.done)
	}()
}

func (m *Manager) StopDisplay() {
	if m.program != nil {
		m.program.Quit()
		if m.done != nil {
			<-m.done
		}
	}
}
