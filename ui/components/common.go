package components

import (
	"github.com/r8bert/rego/ui/styles"
)

type Confirm struct {
	title   string
	message string
	focused int // 0 = yes, 1 = no
}

func NewConfirm(title, message string) *Confirm {
	return &Confirm{title: title, message: message, focused: 1}
}

func (c *Confirm) Left()           { c.focused = 0 }
func (c *Confirm) Right()          { c.focused = 1 }
func (c *Confirm) Toggle()         { c.focused = 1 - c.focused }
func (c *Confirm) Confirmed() bool { return c.focused == 0 }

func (c *Confirm) View() string {
	view := styles.TitleStyle.Render(c.title) + "\n\n"
	view += styles.DescriptionStyle.Render(c.message) + "\n\n"

	yesStyle, noStyle := styles.DimStyle, styles.DimStyle
	if c.focused == 0 {
		yesStyle = styles.SuccessStyle.Bold(true)
	}
	if c.focused == 1 {
		noStyle = styles.ErrorStyle.Bold(true)
	}

	view += "  " + yesStyle.Render("[ Yes ]") + "   " + noStyle.Render("[ No ]")
	return styles.BoxStyle.Render(view)
}

type Header struct {
	title    string
	subtitle string
}

func NewHeader(title, subtitle string) *Header {
	return &Header{title: title, subtitle: subtitle}
}

func (h *Header) View() string {
	return styles.RenderLogo() + "\n" + styles.RenderTagline() + "\n"
}

type Footer struct {
	help string
}

func NewFooter(help string) *Footer { return &Footer{help: help} }
func (f *Footer) SetHelp(h string)  { f.help = h }
func (f *Footer) View() string      { return styles.FooterStyle.Render(f.help) }
