package components

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r8bert/rego/ui/styles"
)

// Spinner frames
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var dotsFrames = []string{"   ", ".  ", ".. ", "..."}
var pulseFrames = []string{"●", "◉", "○", "◉"}
var arrowFrames = []string{"→", "↘", "↓", "↙", "←", "↖", "↑", "↗"}

// TickMsg is sent on each animation frame
type TickMsg time.Time

// Tick returns a command that sends TickMsg
func Tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Spinner is an animated loading indicator
type Spinner struct {
	frame  int
	frames []string
	Style  string
}

func NewSpinner() *Spinner {
	return &Spinner{frames: spinnerFrames, Style: "spinner"}
}

func NewDotsSpinner() *Spinner {
	return &Spinner{frames: dotsFrames, Style: "dots"}
}

func NewPulseSpinner() *Spinner {
	return &Spinner{frames: pulseFrames, Style: "pulse"}
}

func (s *Spinner) Tick() {
	s.frame = (s.frame + 1) % len(s.frames)
}

func (s *Spinner) View() string {
	return styles.WarningStyle.Render(s.frames[s.frame])
}

// AnimatedProgress shows an animated progress bar
type AnimatedProgress struct {
	current   int
	total     int
	width     int
	frame     int
	label     string
	status    string
	startTime time.Time
}

func NewAnimatedProgress(total int) *AnimatedProgress {
	return &AnimatedProgress{total: total, width: 30, startTime: time.Now()}
}

func (p *AnimatedProgress) SetCurrent(c int)   { p.current = c }
func (p *AnimatedProgress) SetLabel(l string)  { p.label = l }
func (p *AnimatedProgress) SetStatus(s string) { p.status = s }
func (p *AnimatedProgress) Tick()              { p.frame++ }
func (p *AnimatedProgress) Increment() {
	if p.current < p.total {
		p.current++
	}
}
func (p *AnimatedProgress) IsComplete() bool { return p.current >= p.total }

func (p *AnimatedProgress) Percent() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total)
}

func (p *AnimatedProgress) View() string {
	var b strings.Builder

	if p.label != "" {
		b.WriteString(styles.SubtitleStyle.Render(p.label) + "\n")
	}

	// Animated progress bar with moving highlight
	filled := int(p.Percent() * float64(p.width))
	empty := p.width - filled

	// Create animated fill effect
	var bar strings.Builder
	for i := 0; i < filled; i++ {
		if i == filled-1 && !p.IsComplete() {
			// Animate the leading edge
			chars := []string{"▓", "▒", "░"}
			bar.WriteString(styles.SuccessStyle.Render(chars[p.frame%len(chars)]))
		} else {
			bar.WriteString(styles.SuccessStyle.Render("█"))
		}
	}
	bar.WriteString(styles.DimStyle.Render(strings.Repeat("░", empty)))

	b.WriteString("[" + bar.String() + "] ")

	// Percentage with animation
	pct := int(p.Percent() * 100)
	b.WriteString(styles.NormalStyle.Render(strings.Repeat(" ", 3-len(string(rune('0'+pct%10)))) + string(rune('0'+pct/10)) + string(rune('0'+pct%10)) + "%"))
	b.WriteString("\n")

	// Status with spinner
	if p.status != "" && !p.IsComplete() {
		spinner := spinnerFrames[p.frame%len(spinnerFrames)]
		b.WriteString(styles.WarningStyle.Render(spinner) + " " + styles.StatusStyle.Render(p.status) + "\n")
	}

	// Elapsed time
	elapsed := time.Since(p.startTime).Round(time.Second)
	b.WriteString(styles.DimStyle.Render("Elapsed: "+elapsed.String()) + "\n")

	return b.String()
}

// AnimatedText shows text with typing effect
type AnimatedText struct {
	fullText string
	visible  int
	speed    int // chars per tick
	done     bool
}

func NewAnimatedText(text string) *AnimatedText {
	return &AnimatedText{fullText: text, speed: 2}
}

func (t *AnimatedText) Tick() {
	if t.visible < len(t.fullText) {
		t.visible += t.speed
		if t.visible > len(t.fullText) {
			t.visible = len(t.fullText)
			t.done = true
		}
	}
}

func (t *AnimatedText) IsDone() bool { return t.done }
func (t *AnimatedText) Skip()        { t.visible = len(t.fullText); t.done = true }

func (t *AnimatedText) View() string {
	return t.fullText[:t.visible]
}

// Blink creates a blinking effect
type Blink struct {
	text    string
	visible bool
	frame   int
}

func NewBlink(text string) *Blink {
	return &Blink{text: text, visible: true}
}

func (b *Blink) Tick() {
	b.frame++
	b.visible = (b.frame/5)%2 == 0
}

func (b *Blink) View() string {
	if b.visible {
		return b.text
	}
	return strings.Repeat(" ", len(b.text))
}

// PulseText creates a pulsing color effect
type PulseText struct {
	text   string
	frame  int
	colors []string
}

func NewPulseText(text string) *PulseText {
	return &PulseText{
		text:   text,
		colors: []string{"#7C3AED", "#9333EA", "#A855F7", "#9333EA"},
	}
}

func (p *PulseText) Tick() { p.frame++ }

func (p *PulseText) View() string {
	// Simple pulse - just return styled text
	return styles.SelectedStyle.Render(p.text)
}
