package components

import (
	"strings"

	"github.com/r8bert/rego/ui/styles"
)

type Progress struct {
	current int
	total   int
	width   int
	label   string
	status  string
}

func NewProgress(total int) *Progress {
	return &Progress{total: total, width: 40}
}

func (p *Progress) SetWidth(w int)     { p.width = w }
func (p *Progress) SetCurrent(c int)   { p.current = c }
func (p *Progress) SetLabel(l string)  { p.label = l }
func (p *Progress) SetStatus(s string) { p.status = s }
func (p *Progress) Increment() {
	if p.current < p.total {
		p.current++
	}
}
func (p *Progress) Percent() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total)
}
func (p *Progress) IsComplete() bool { return p.current >= p.total }

func (p *Progress) View() string {
	var b strings.Builder

	if p.label != "" {
		b.WriteString(styles.SubtitleStyle.Render(p.label) + "\n")
	}

	filled := int(p.Percent() * float64(p.width))
	empty := p.width - filled
	bar := styles.ProgressBarStyle.Render(strings.Repeat("█", filled)) + styles.DimStyle.Render(strings.Repeat("░", empty))
	b.WriteString("[" + bar + "] ")
	b.WriteString(styles.NormalStyle.Render(string(rune('0'+p.current%10)) + "/" + string(rune('0'+p.total%10))))
	b.WriteString("\n")

	if p.status != "" {
		b.WriteString(styles.StatusStyle.Render(p.status) + "\n")
	}

	return b.String()
}

// StatusList shows operation results
type StatusItem struct {
	Label   string
	Status  string // "success", "error", "pending", "running"
	Message string
}

type StatusList struct {
	items []StatusItem
}

func NewStatusList() *StatusList          { return &StatusList{} }
func (s *StatusList) Add(item StatusItem) { s.items = append(s.items, item) }
func (s *StatusList) Clear()              { s.items = nil }

func (s *StatusList) View() string {
	var b strings.Builder
	for _, item := range s.items {
		var icon, style string
		switch item.Status {
		case "success":
			icon = "✓"
			style = styles.SuccessStyle.Render(icon)
		case "error":
			icon = "✗"
			style = styles.ErrorStyle.Render(icon)
		case "pending":
			icon = "○"
			style = styles.DimStyle.Render(icon)
		case "running":
			icon = "◉"
			style = styles.WarningStyle.Render(icon)
		default:
			icon = "•"
			style = styles.NormalStyle.Render(icon)
		}
		b.WriteString("  " + style + " " + item.Label)
		if item.Message != "" {
			b.WriteString(" - " + styles.DimStyle.Render(item.Message))
		}
		b.WriteString("\n")
	}
	return b.String()
}
