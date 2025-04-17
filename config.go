package main

import (
	db "github.com/Asfolny/protastim/internal/database"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type config struct {
	lg       *lipgloss.Renderer
	styles   *styles
	maxWidth int
	queries  *db.Queries
	size     tea.WindowSizeMsg
}

func newConfig(q *db.Queries) *config {
	lg := lipgloss.DefaultRenderer()
	styles := newStyles(lg)

	return &config{
		lg: lg,
		styles: styles,
		maxWidth: maxWidth,
		queries: q,
		size: tea.WindowSizeMsg{ // Default size
			Height: 60,
			Width: 80,
		},
	}
}

const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type styles struct {
	Base,
	BaseTable,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help,
	Spinner lipgloss.Style
}

// The outer box is defined as 2 on either side, see wrapperStyle
func (c *config) getInnerWidth() int {
	return c.size.Width - 4
}

// The top bar is 1 line, and the wrapperStyle adds -2, so the most innerHeight has to be - 3
func (c *config) getInnerHeight() int {
	return c.size.Height - 3
}

func (c *config) newSpinner() spinner.Model {
	return spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(c.styles.Spinner),
	)
}

func newStyles(lg *lipgloss.Renderer) *styles {
	s := styles{}

	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)

	s.BaseTable = lg.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63"))

	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)

	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)

	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)

	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))

	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)

	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))

	s.Spinner = lg.NewStyle().
		Foreground(lipgloss.Color("69"))

	return &s
}
