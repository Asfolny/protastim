package main

import (
	"database/sql"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

type config struct {
	maxWidth int
	queries  *db.Queries
	db       *sql.DB
	size     tea.WindowSizeMsg
}

func newConfig(conn *sql.DB) *config {
	return &config{
		maxWidth: maxWidth,
		queries: db.New(conn),
		db: conn,
		size: tea.WindowSizeMsg{
			Height: 60,
			Width: 80,
		},
	}
}

const maxWidth = 80

func (c *config) getInnerWidth() int {
	return c.size.Width - 2
}

func (c *config) getInnerHeight() int {
	return c.size.Height - 2
}
