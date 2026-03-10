package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var Version = "dev"

type model struct {
	r, g, b float64
	cursor  int
	msg     string
}

func initialModel() model {
	return model{r: 0.5, g: 0.2, b: 0.8, cursor: 0}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.msg = ""
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 { m.cursor-- }
		case "down", "j":
			if m.cursor < 2 { m.cursor++ }
		case "left", "h":
			m.adjustColor(-0.01)
		case "right", "l":
			m.adjustColor(0.01)
		case "L":
			m.adjustColor(0.05)
		case "H":
			m.adjustColor(-0.05)
		case "c": // Copy Hex
			c := colorful.Color{R: m.r, G: m.g, B: m.b}
			clipboard.WriteAll(c.Hex())
			m.msg = "Hex copied!"
		}
	}
	return m, nil
}

func (m *model) adjustColor(amount float64) {
	switch m.cursor {
	case 0: m.r = clamp(m.r + amount)
	case 1: m.g = clamp(m.g + amount)
	case 2: m.b = clamp(m.b + amount)
	}
}

func clamp(v float64) float64 {
	if v < 0 { return 0 }
	if v > 1 { return 1 }
	return v
}

func getWCAG(c colorful.Color) string {
	white := colorful.Color{R: 1, G: 1, B: 1}
	black := colorful.Color{R: 0, G: 0, B: 0}
	ratioWhite := c.DistanceLuv(white)
	ratioBlack := c.DistanceLuv(black)
	
	if ratioWhite > ratioBlack {
		return "Best with White text"
	}
	return "Best with Black text"
}

func (m model) View() string {
	c := colorful.Color{R: m.r, G: m.g, B: m.b}
	hex := c.Hex()
	h, s, l := c.Hsl()
	r255, g255, b255 := int(m.r*255), int(m.g*255), int(m.b*255)

	// CMYK
	cyan, magenta, yellow, black := rgbToCmyk(m.r, m.g, m.b)

	sStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Background(lipgloss.Color(hex)).Padding(1, 4).MarginRight(2)
	infoStyle := lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Atlas Color Pro") + "\n\n")
	
	preview := sStyle.Render(" ")
	details := fmt.Sprintf(
		"HEX:  %s\nRGB:  %d, %d, %d\nHSL:  %.0f°, %.0f%%, %.0f%%\nCMYK: %.0f%%, %.0f%%, %.0f%%, %.0f%%\nWCAG: %s",
		hex, r255, g255, b255, h, s*100, l*100, cyan*100, magenta*100, yellow*100, black*100, getWCAG(c),
	)
	
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, preview, infoStyle.Render(details)) + "\n\n")

	// Sliders
	rows := []string{"Red", "Green", "Blue"}
	vals := []float64{m.r, m.g, m.b}
	for i, name := range rows {
		cursor := " "; if m.cursor == i { cursor = ">" }
		barWidth := 30
		filled := int(vals[i] * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		
		colorCode := "255"
		if i == 0 { colorCode = "196" }
		if i == 1 { colorCode = "46" }
		if i == 2 { colorCode = "21" }

		rowStyle := lipgloss.NewStyle()
		if m.cursor == i { rowStyle = rowStyle.Foreground(lipgloss.Color("63")).Bold(true) }
		b.WriteString(fmt.Sprintf("%s %-5s [%s] %.0f%%\n", cursor, rowStyle.Render(name), lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode)).Render(bar), vals[i]*100))
	}

	// Palettes
	b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("Harmonies:") + "\n")
	
	comp := c; hComp, sComp, lComp := comp.Hsl(); hComp = mod(hComp+180, 360)
	comp = colorful.Hsl(hComp, sComp, lComp)
	
	tri1 := colorful.Hsl(mod(h+120, 360), s, l)
	tri2 := colorful.Hsl(mod(h+240, 360), s, l)

	b.WriteString(fmt.Sprintf("Comp: %s  Triadic: %s, %s\n", 
		lipgloss.NewStyle().Foreground(lipgloss.Color(comp.Hex())).Render("■"),
		lipgloss.NewStyle().Foreground(lipgloss.Color(tri1.Hex())).Render("■"),
		lipgloss.NewStyle().Foreground(lipgloss.Color(tri2.Hex())).Render("■"),
	))

	if m.msg != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(m.msg))
	}

	b.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("c: copy hex • arrows/hjkl: adjust • q: quit"))
	return b.String()
}

func mod(a, b float64) float64 {
	for a >= b { a -= b }
	for a < 0 { a += b }
	return a
}

func rgbToCmyk(r, g, b float64) (c, m, y, k float64) {
	k = 1.0 - max(r, max(g, b))
	if k == 1.0 {
		return 0, 0, 0, 1.0
	}
	c = (1.0 - r - k) / (1.0 - k)
	m = (1.0 - g - k) / (1.0 - k)
	y = (1.0 - b - k) / (1.0 - k)
	return
}

func max(a, b float64) float64 {
	if a > b { return a }
	return b
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("atlas.color v%s\n", Version)
		return
	}
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
