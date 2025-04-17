package main

import (
	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

type config struct {
	maxWidth int
	queries  *db.Queries
	size     tea.WindowSizeMsg
}

func newConfig(q *db.Queries) *config {
	return &config{
		maxWidth: maxWidth,
		queries: q,
		size: tea.WindowSizeMsg{
			Height: 60,
			Width: 80,
		},
	}
}

const maxWidth = 80

// The outer box is defined as 2 on either side, see wrapperStyle
func (c *config) getInnerWidth() int {
	return c.size.Width - 4
}

// The top bar is 1 line, and the wrapperStyle adds -2, so the most innerHeight has to be - 3
func (c *config) getInnerHeight() int {
	return c.size.Height - 3
}
