package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

type FunctionOutput struct {
	Name        string
	Status      string
	Message     string
	StreamLines []string
	Complete    bool
	StartTime   time.Time
	LastUpdated time.Time
	Error       error
	Index       int
}

type ErrorReport struct {
	FunctionName string
	Error        error
	Time         time.Time
}

type Manager struct {
	outputs       map[string]*FunctionOutput
	mutex         sync.RWMutex
	noClear       bool
	numLines      int
	maxStreams    int               // Max output stream lines per function
	tables        map[string]*Table // Global tables
	errors        []ErrorReport
	doneCh        chan struct{} // Channel to signal stopping the display
	pauseCh       chan bool     // Channel to pause/resume display updates
	isPaused      bool
	displayTick   time.Duration // Interval between display updates
	functionCount int
	displayWg     sync.WaitGroup // WaitGroup for display goroutine shutdown
}

var GlobalDebugFlag bool

func NewManager() *Manager {
	retMgr := &Manager{
		outputs:       make(map[string]*FunctionOutput),
		noClear:       false,
		tables:        make(map[string]*Table),
		errors:        []ErrorReport{},
		maxStreams:    15,
		doneCh:        make(chan struct{}),
		pauseCh:       make(chan bool),
		isPaused:      false,
		displayTick:   200 * time.Millisecond, // Default
		functionCount: 0,
	}
	if GlobalDebugFlag {
		retMgr.displayTick = 1 * time.Second // slow refresh for debug mode
		retMgr.noClear = true
	}
	return retMgr
}

func (m *Manager) Pause() {
	if !m.isPaused {
		m.pauseCh <- true
		m.isPaused = true
	}
}

func (m *Manager) Resume() {
	if m.isPaused {
		m.pauseCh <- false
		m.isPaused = false
	}
}

func (m *Manager) Register(name string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.functionCount++
	m.outputs[fmt.Sprint(m.functionCount)] = &FunctionOutput{
		Name:        name,
		Status:      "pending",
		StreamLines: []string{},
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
		Index:       m.functionCount,
	}
	return fmt.Sprint(m.functionCount)
}

func (m *Manager) SetMessage(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Message = message
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) Complete(name, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.StreamLines = []string{}
		if message == "" {
			info.Message = fmt.Sprintf("Completed %s", info.Name)
		} else {
			info.Message = message
		}
		info.Complete = true
		info.Status = "success"
		info.LastUpdated = time.Now()
	}
}

func (m *Manager) ReportError(name string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		info.Complete = true
		info.Status = "error"
		info.Error = err
		info.LastUpdated = time.Now()
		// Add to global error list
		m.errors = append(m.errors, ErrorReport{
			FunctionName: name,
			Error:        err,
			Time:         time.Now(),
		})
	}
}

// func (m *Manager) UpdateStreamOutput(name string, output []string) {
// 	m.mutex.Lock()
// 	defer m.mutex.Unlock()
// 	if info, exists := m.outputs[name]; exists {
// 		currentLen := len(info.StreamLines)
// 		if currentLen+len(output) > m.maxStreams {
// 			startIndex := currentLen + len(output) - m.maxStreams
// 			if startIndex > currentLen {
// 				startIndex = 0
// 			}
// 			newLines := append(info.StreamLines[startIndex:], output...)
// 			if len(newLines) > m.maxStreams {
// 				newLines = newLines[len(newLines)-m.maxStreams:]
// 			}
// 			info.StreamLines = newLines
// 		} else {
// 			info.StreamLines = append(info.StreamLines, output...)
// 		}
// 		info.LastUpdated = time.Now()
// 	}
// }

func (m *Manager) AddStreamLine(name string, line string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		wrappedLines := wrapText(line, basePadding+4)
		currentLen := len(info.StreamLines)
		totalNewLines := len(wrappedLines)
		if currentLen+totalNewLines > m.maxStreams {
			startIndex := currentLen + totalNewLines - m.maxStreams
			if startIndex > currentLen {
				startIndex = 0
				existingToKeep := m.maxStreams - totalNewLines
				if existingToKeep > 0 {
					info.StreamLines = info.StreamLines[currentLen-existingToKeep:]
				} else {
					info.StreamLines = []string{} // All existing lines will be dropped
				}
			} else {
				info.StreamLines = info.StreamLines[startIndex:]
			}
			info.StreamLines = append(info.StreamLines, wrappedLines...)
		} else {
			info.StreamLines = append(info.StreamLines, wrappedLines...)
		}
		if len(info.StreamLines) > m.maxStreams {
			info.StreamLines = info.StreamLines[len(info.StreamLines)-m.maxStreams:]
		}
		info.LastUpdated = time.Now()
	}
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Default fallback width if terminal width can't be determined
	}
	return width
}

func wrapText(text string, indent int) []string {
	termWidth := getTerminalWidth()
	maxWidth := termWidth - indent - 2 // Account for indentation
	if maxWidth <= 10 {
		maxWidth = 80
	}
	if utf8.RuneCountInString(text) <= maxWidth {
		return []string{text}
	}
	var lines []string
	currentLine := ""
	currentWidth := 0
	for _, r := range text {
		runeWidth := 1
		// If adding this rune would exceed max width, flush the line
		if currentWidth+runeWidth > maxWidth {
			lines = append(lines, currentLine)
			currentLine = string(r)
			currentWidth = runeWidth
		} else {
			currentLine += string(r)
			currentWidth += runeWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func (m *Manager) AddProgressBarToStream(name string, outof, final int64, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.outputs[name]; exists {
		progressBar := PrintProgressBar(outof, final, 30)
		display := progressBar + debugStyle.Render(text)
		info.StreamLines = []string{display} // Set as only stream so nothing else is displayed
		info.LastUpdated = time.Now()
	}
}

func PrintProgressBar(current, total int64, width int) string {
	if width <= 0 {
		width = 30
	}
	percent := float64(current) / float64(total)
	filled := min(int(percent*float64(width)), width)
	bar := StyleSymbols["bullet"]
	bar += strings.Repeat(StyleSymbols["hline"], filled)
	if filled < width {
		bar += strings.Repeat(" ", width-filled)
	}
	bar += StyleSymbols["bullet"]
	return debugStyle.Render(fmt.Sprintf("%s %.1f%% %s ", bar, percent*100, StyleSymbols["bullet"]))
}

func (m *Manager) ClearAll() {
	if m.noClear {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for name := range m.outputs {
		m.outputs[name].StreamLines = []string{}
	}
}

func (m *Manager) GetStatusIndicator(status string) string {
	switch status {
	case "success", "pass":
		return successStyle.Render(StyleSymbols["pass"])
	case "error", "fail":
		return errorStyle.Render(StyleSymbols["fail"])
	case "warning":
		return warningStyle.Render(StyleSymbols["warning"])
	case "pending":
		return pendingStyle.Render(StyleSymbols["pending"])
	default:
		return infoStyle.Render(StyleSymbols["bullet"])
	}
}

// Add a global table
func (m *Manager) RegisterTable(name string, headers []string) *Table {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	table := NewTable(headers)
	m.tables[name] = table
	return table
}

func (m *Manager) sortFunctions() (active, pending, completed []*FunctionOutput) {
	var allFuncs []*FunctionOutput
	// Sort by index (registration order)
	for _, info := range m.outputs {
		allFuncs = append(allFuncs, info)
	}
	sort.Slice(allFuncs, func(i, j int) bool {
		return allFuncs[i].Index < allFuncs[j].Index
	})
	// Group functions by status
	for _, f := range allFuncs {
		if f.Complete {
			completed = append(completed, f)
		} else if f.Status == "pending" && f.Message == "" {
			pending = append(pending, f)
		} else {
			active = append(active, f)
		}
	}
	return active, pending, completed
}

func (m *Manager) updateDisplay() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.numLines > 0 && !m.noClear {
		fmt.Printf("\033[%dA\033[J", m.numLines)
	}
	lineCount := 0
	activeFuncs, pendingFuncs, completedFuncs := m.sortFunctions()

	// Display active functions
	for _, f := range activeFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		elapsed := time.Since(info.StartTime).Round(time.Second)
		if info.Complete {
			elapsed = info.LastUpdated.Sub(info.StartTime).Round(time.Second)
		}
		elapsedStr := elapsed.String()

		// Style the message based on status
		var styledMessage string
		switch info.Status {
		case "success":
			styledMessage = successStyle.Render(info.Message)
		case "error":
			styledMessage = errorStyle.Render(info.Message)
		case "warning":
			styledMessage = warningStyle.Render(info.Message)
		default: // pending or other
			styledMessage = pendingStyle.Render(info.Message)
		}
		fmt.Printf("%s%s %s %s\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(elapsedStr), styledMessage)
		lineCount++

		// Print stream lines with indentation
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4) // Additional indentation for stream output
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display pending functions
	for _, f := range pendingFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		fmt.Printf("%s%s %s\n", strings.Repeat(" ", basePadding), statusDisplay, pendingStyle.Render("Waiting..."))
		lineCount++
		if len(info.StreamLines) > 0 {
			indent := strings.Repeat(" ", basePadding+4)
			for _, line := range info.StreamLines {
				fmt.Printf("%s%s\n", indent, streamStyle.Render(line))
				lineCount++
			}
		}
	}

	// Display completed functions
	for _, f := range completedFuncs {
		info := f
		statusDisplay := m.GetStatusIndicator(info.Status)
		totalTime := info.LastUpdated.Sub(info.StartTime).Round(time.Second)
		timeStr := totalTime.String()

		var styledMessage string
		switch info.Status {
		case "success":
			styledMessage = successStyle.Render(info.Message)
		case "error":
			styledMessage = errorStyle.Render(info.Message)
		case "warning":
			styledMessage = warningStyle.Render(info.Message)
		default: // pending or other
			styledMessage = pendingStyle.Render(info.Message)
		}
		fmt.Printf("%s%s %s %s\n", strings.Repeat(" ", basePadding), statusDisplay, debugStyle.Render(timeStr), styledMessage)
		lineCount++
	}
	m.numLines = lineCount
}

func (m *Manager) StartDisplay() {
	m.displayWg.Add(1)
	go func() {
		defer m.displayWg.Done()
		ticker := time.NewTicker(m.displayTick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !m.isPaused {
					m.updateDisplay()
				}
			case pauseState := <-m.pauseCh:
				m.isPaused = pauseState
			case <-m.doneCh:
				m.ClearAll()
				m.updateDisplay()
				m.ShowSummary()
				m.displayTables()
				return
			}
		}
	}()
}

func (m *Manager) StopDisplay() {
	close(m.doneCh)
	m.displayWg.Wait() // Wait for goroutine to finish
}

func (m *Manager) displayTables() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.tables) > 0 {
		for name, table := range m.tables {
			if len(table.Rows) == 0 {
				continue
			}
			fmt.Println(strings.Repeat(" ", basePadding) + infoStyle.Render(name))
			fmt.Println(table.FormatTable(false))
		}
	}
}

func (m *Manager) displayErrors() {
	if len(m.errors) == 0 {
		return
	}
	fmt.Println()
	fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Bold(true).Render("Errors:"))
	for i, err := range m.errors {
		fmt.Printf("%s%s %s %s\n",
			strings.Repeat(" ", basePadding+2),
			errorStyle.Render(fmt.Sprintf("%d.", i+1)),
			debugStyle.Render(fmt.Sprintf("[%s]", err.Time.Format("15:04:05"))),
			errorStyle.Render(fmt.Sprintf("Function: %s", err.FunctionName)))
		fmt.Printf("%s%s\n", strings.Repeat(" ", basePadding+4), errorStyle.Render(fmt.Sprintf("Error: %v", err.Error)))
	}
}

func (m *Manager) ShowSummary() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	fmt.Println()
	var success, failures int
	for _, info := range m.outputs {
		switch info.Status {
		case "success":
			success++
		case "error":
			failures++
		}
	}
	failed := fmt.Sprintf("Failed %d of %d", failures, len(m.outputs))
	if failures > 0 {
		fmt.Println(strings.Repeat(" ", basePadding) + errorStyle.Render(failed))
	}
	m.displayErrors()
	fmt.Println()
}
