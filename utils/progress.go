package utils

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/rs/zerolog/log"
)

type Unit string

const UnitBytes Unit = ""

const (
	minBarWidth        = 8
	maxBarWidth        = 30
	minWidth           = 24
	defaultWidth       = 80
	meterIndent        = 2
	fieldGap           = 2
	reservePercent     = 4
	reserveTransferred = 14
	reserveRate        = 11
	reserveETA         = 11
	reserveAvg         = 15
	rateWindowSpan     = 800 * time.Millisecond
	rateFloor          = 200 * time.Millisecond
	maxETA             = 100 * time.Hour
)

var (
	barFilled = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(12))
	barEmpty  = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))
	mutedText = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))
)

var (
	cursorOnce   sync.Once
	cursorMu     sync.Mutex
	cursorHidden bool
)

type sample struct {
	at  time.Time
	val int64
}

type rateWindow struct {
	samples []sample
}

func (r *rateWindow) add(val int64, at time.Time) {
	r.samples = append(r.samples, sample{at: at, val: val})
	cutoff := at.Add(-rateWindowSpan)
	i := 0
	for i < len(r.samples) && r.samples[i].at.Before(cutoff) {
		i++
	}
	if i > 0 {
		r.samples = r.samples[i:]
	}
}

func (r *rateWindow) current() float64 {
	if len(r.samples) < 2 {
		return 0
	}
	first, last := r.samples[0], r.samples[len(r.samples)-1]
	if last.at.Sub(first.at) < rateFloor {
		return 0
	}
	delta := float64(last.val - first.val)
	if delta <= 0 {
		return 0
	}
	return delta / last.at.Sub(first.at).Seconds()
}

type Meter struct {
	verb    string
	name    string
	context string
	total   int64
	unit    Unit

	mu      sync.Mutex
	current int64
	start   time.Time
	window  rateWindow
	ticker  *time.Ticker
	stopCh  chan struct{}
	drawn   int
	settled bool
}

func NewMeter(verb, name string, total int64, unit Unit) *Meter {
	interval := time.Second
	if StdoutIsTerminal && !GlobalDebugFlag {
		interval = 100 * time.Millisecond
	}
	m := &Meter{
		verb:   verb,
		name:   name,
		total:  total,
		unit:   unit,
		start:  time.Now(),
		ticker: time.NewTicker(interval),
		stopCh: make(chan struct{}),
	}
	if m.live() {
		hideCursor()
		m.draw(m.frame())
	}
	go m.loop()
	return m
}

func (m *Meter) live() bool {
	return StdoutIsTerminal && !GlobalDebugFlag
}

func (m *Meter) Context(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.context = s
}

func (m *Meter) Write(p []byte) (int, error) {
	m.Add(int64(len(p)))
	return len(p), nil
}

func (m *Meter) Add(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.settled {
		return
	}
	m.current += n
	if m.current < 0 {
		m.current = 0
	}
	now := time.Now()
	m.window.add(m.current, now)
}

func (m *Meter) Done() {
	m.mu.Lock()
	if m.settled {
		m.mu.Unlock()
		return
	}
	m.finishLocked()
	elapsed := time.Since(m.start)
	cur := m.current
	name := m.name
	unit := m.unit
	m.mu.Unlock()

	line := settledLine(name, cur, elapsed, unit)
	PrintSuccess(line)
}

func (m *Meter) Fail(err error) {
	m.mu.Lock()
	if m.settled {
		m.mu.Unlock()
		return
	}
	name := m.name
	m.finishLocked()
	m.mu.Unlock()

	if GlobalDebugFlag {
		log.Error().Err(err).Str("name", name).Msg("failed")
		return
	}
	msg := name
	if err != nil {
		msg = name + ": " + err.Error()
	}
	lipgloss.Println(errorStyle.Render("  " + StyleSymbols["fail"] + " " + msg))
}

func (m *Meter) finishLocked() {
	m.settled = true
	if m.ticker != nil {
		m.ticker.Stop()
	}
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
	if m.live() {
		m.draw(nil)
	}
	showCursor()
}

func (m *Meter) loop() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.ticker.C:
			m.mu.Lock()
			if m.settled {
				m.mu.Unlock()
				return
			}
			m.tickLocked()
			m.mu.Unlock()
		}
	}
}

func (m *Meter) tickLocked() {
	if m.live() {
		m.draw(m.frame())
		return
	}
	if GlobalDebugFlag {
		m.debugTick()
		return
	}
	if m.current > 0 {
		lipgloss.Println(m.pipedLine())
	}
}

func (m *Meter) draw(lines []string) {
	var b strings.Builder
	b.WriteString(strings.Repeat("\033[1A\033[2K", m.drawn))
	for _, l := range lines {
		b.WriteString("\r\033[2K" + l + "\n")
	}
	fmt.Print(b.String())
	m.drawn = len(lines)
}

func (m *Meter) frame() []string {
	return []string{m.headerLine(termWidth()), m.meterLine(termWidth())}
}

func (m *Meter) headerLine(width int) string {
	glyph := StyleSymbols["running"] + " "
	verb := m.verb
	name := m.name
	ctx := m.context
	budget := width - displayLen(glyph)
	if budget < 1 {
		return infoStyle.Render(glyph)
	}
	prefix := verb + " "
	rest := budget - displayLen(prefix)
	if rest < 1 {
		return infoStyle.Render(glyph + clip(verb, budget))
	}
	ctxPart := ""
	if ctx != "" {
		ctxPart = " " + ctx
	}
	if displayLen(name)+displayLen(ctxPart) > rest {
		ctxPart = ""
	}
	name = clip(name, rest-displayLen(ctxPart))
	line := infoStyle.Render(glyph+prefix+name) + mutedText.Render(ctxPart)
	return line
}

func (m *Meter) meterLine(width int) string {
	pct := ""
	if m.total > 0 {
		p := int(float64(m.current) * 100 / float64(m.total))
		if p > 100 {
			p = 100
		}
		if p < 0 {
			p = 0
		}
		pct = fmt.Sprintf("%3d%%", p)
	}
	transferred := formatPair(m.current, m.total, m.unit)
	elapsed := time.Since(m.start)
	rate := m.window.current()
	avg := m.average(elapsed, true)
	rateStr := formatRate(rate, m.unit)
	etaStr := formatETA(m.total, m.current, rate)
	avgStr := "avg " + formatRate(avg, m.unit)

	fixed := make([]meterField, 0, 2)
	if pct != "" {
		fixed = append(fixed, meterField{pct, reservePercent})
	}
	fixed = append(fixed, meterField{transferred, reserveTransferred})
	optionals := []meterField{
		{rateStr, reserveRate},
		{etaStr, reserveETA},
		{avgStr, reserveAvg},
	}

	for drop := 0; drop <= len(optionals); drop++ {
		fields := append(append([]meterField{}, fixed...), optionals[:len(optionals)-drop]...)
		reserved := 0
		for _, f := range fields {
			reserved += f.reserved
		}
		barBudget := width - meterIndent - reserved - fieldGap*len(fields)
		if barBudget >= minBarWidth {
			barW := min(maxBarWidth, barBudget)
			parts := make([]string, 0, 1+len(fields))
			parts = append(parts, m.bar(barW))
			for _, f := range fields {
				parts = append(parts, f.text)
			}
			return strings.Repeat(" ", meterIndent) + strings.Join(parts, "  ")
		}
	}
	for drop := 0; drop <= len(optionals); drop++ {
		fields := append(append([]meterField{}, fixed...), optionals[:len(optionals)-drop]...)
		parts := make([]string, 0, len(fields))
		for _, f := range fields {
			parts = append(parts, f.text)
		}
		line := strings.Repeat(" ", meterIndent) + strings.Join(parts, "  ")
		if displayLen(line) <= width {
			return line
		}
	}
	parts := make([]string, 0, len(fixed))
	for _, f := range fixed {
		parts = append(parts, f.text)
	}
	return strings.Repeat(" ", meterIndent) + strings.Join(parts, "  ")
}

type meterField struct {
	text     string
	reserved int
}

func (m *Meter) bar(width int) string {
	if width < minBarWidth {
		return ""
	}
	if m.total <= 0 {
		blob := 4
		if blob > width {
			blob = width
		}
		shift := int(time.Since(m.start).Milliseconds()/80) % (width + blob)
		var b strings.Builder
		for i := range width {
			if i >= shift-blob && i < shift {
				b.WriteString(barFilled.Render("─"))
			} else {
				b.WriteString(barEmpty.Render("─"))
			}
		}
		return b.String()
	}
	filled := int(float64(width) * float64(m.current) / float64(m.total))
	filled = min(max(filled, 0), width)
	return barFilled.Render(strings.Repeat("─", filled)) + barEmpty.Render(strings.Repeat("─", width-filled))
}

func (m *Meter) pipedLine() string {
	parts := []string{StyleSymbols["running"] + " " + m.verb + " " + m.name}
	if m.total > 0 {
		p := int(float64(m.current) * 100 / float64(m.total))
		parts = append(parts, fmt.Sprintf("%d%%", p), formatPair(m.current, m.total, m.unit))
	} else {
		parts = append(parts, formatPair(m.current, m.total, m.unit))
	}
	rate := m.window.current()
	if rate > 0 {
		parts = append(parts, formatRate(rate, m.unit))
	}
	return strings.Join(parts, "  ")
}

func (m *Meter) debugTick() {
	pct := 0
	if m.total > 0 {
		pct = int(float64(m.current) * 100 / float64(m.total))
	}
	rate := m.window.current()
	eta := formatETA(m.total, m.current, rate)
	log.Info().
		Int("percent", pct).
		Int64("current", m.current).
		Int64("total", m.total).
		Float64("rate", rate).
		Str("eta", eta).
		Msg(m.verb + " " + m.name)
}

func (m *Meter) average(elapsed time.Duration, live bool) float64 {
	if elapsed <= 0 || m.current <= 0 {
		return 0
	}
	if live && elapsed < rateFloor {
		return 0
	}
	return float64(m.current) / elapsed.Seconds()
}

func settledLine(name string, amount int64, elapsed time.Duration, unit Unit) string {
	name = clip(name, max(termWidth()-40, 8))
	parts := []string{name, formatScalar(amount, unit), formatElapsed(elapsed), "avg " + formatRate(float64(amount)/max(elapsed.Seconds(), 0.0001), unit)}
	if amount <= 0 {
		parts = []string{name, formatElapsed(elapsed)}
	}
	return strings.Join(parts, "  ")
}

func ClearLines(n int) {
	if GlobalDebugFlag || !StdoutIsTerminal {
		return
	}
	for range n {
		fmt.Print("\033[A\033[2K")
	}
}

func ClearPreviousLine() {
	ClearLines(1)
}

func termWidth() int {
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil && w > 0 {
		return w
	}
	if n, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && n >= minWidth {
		return n
	}
	return defaultWidth
}

func hideCursor() {
	cursorOnce.Do(func() {
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c
			showCursor()
			os.Exit(130)
		}()
	})
	cursorMu.Lock()
	defer cursorMu.Unlock()
	if !cursorHidden {
		fmt.Print("\033[?25l")
		cursorHidden = true
	}
}

func showCursor() {
	cursorMu.Lock()
	defer cursorMu.Unlock()
	if cursorHidden {
		fmt.Print("\033[?25h")
		cursorHidden = false
	}
}

func displayLen(s string) int {
	return lipgloss.Width(s)
}

func clip(s string, maxCells int) string {
	if maxCells <= 0 {
		return ""
	}
	if displayLen(s) <= maxCells {
		return s
	}
	if maxCells == 1 {
		return "…"
	}
	r := []rune(s)
	for len(r) > 0 && displayLen(string(r)) > maxCells-1 {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func formatPair(cur, total int64, unit Unit) string {
	if unit != UnitBytes {
		if total <= 0 {
			return fmt.Sprintf("%d %s", cur, unit)
		}
		return fmt.Sprintf("%d / %d %s", cur, total, unit)
	}
	if total <= 0 {
		return formatBytes(cur)
	}
	label, div := byteUnit(total)
	return fmt.Sprintf("%s / %s %s", formatScaled(float64(cur)/div), formatScaled(float64(total)/div), label)
}

func formatScalar(n int64, unit Unit) string {
	if unit != UnitBytes {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return formatBytes(n)
}

func formatRate(bps float64, unit Unit) string {
	if bps <= 0 {
		if unit != UnitBytes {
			return fmt.Sprintf("0.0 %s/s", unit)
		}
		return "0.0 B/s"
	}
	if unit != UnitBytes {
		return fmt.Sprintf("%s %s/s", formatScaled(bps), unit)
	}
	return formatBytes(int64(bps)) + "/s"
}

func formatBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	label, div := byteUnit(n)
	if label == "B" {
		return fmt.Sprintf("%d B", n)
	}
	return formatScaled(float64(n)/div) + " " + label
}

func byteUnit(n int64) (string, float64) {
	const k = 1024.0
	v := float64(max(n, 0))
	switch {
	case v < k:
		return "B", 1
	case v < k*k:
		return "KB", k
	case v < k*k*k:
		return "MB", k * k
	case v < k*k*k*k:
		return "GB", k * k * k
	default:
		return "TB", k * k * k * k
	}
}

func formatScaled(n float64) string {
	if n < 100 {
		return fmt.Sprintf("%.1f", n)
	}
	return fmt.Sprintf("%.0f", n)
}

func formatElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	h := int(d.Hours())
	min := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%02dm", h, min)
}

func formatETA(total, current int64, rate float64) string {
	if total <= 0 || rate <= 0 || current >= total {
		return "eta unknown"
	}
	remain := float64(total-current) / rate
	d := time.Duration(remain * float64(time.Second))
	if d > maxETA {
		return "eta unknown"
	}
	if d < time.Minute {
		return fmt.Sprintf("eta %ds", int(d.Seconds()+0.5))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("eta %dm%02ds", m, s)
	}
	h := int(d.Hours())
	min := int(d.Minutes()) % 60
	return fmt.Sprintf("eta %dh%02dm", h, min)
}
