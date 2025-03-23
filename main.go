package main

import (
	"strings"
	"database/sql"
	"fmt"
	"os"

	db "github.com/Asfolny/protastim/internal/database"
	goose "github.com/Asfolny/protastim/internal/sql"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	view   tea.Model
	config *config
}

func (m model) Init() tea.Cmd {
	return m.view.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case newViewMsg:
		m.view = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.config.styles.HeaderText.Render("Protastim\n\n"))
	b.WriteString(m.view.View())

	return b.String()
}

type newViewMsg tea.Model

func changeView(newView tea.Model) tea.Cmd {
	return tea.Batch(newView.Init(), func() tea.Msg {
		return newViewMsg(newView)
	})
}

func main() {
	// TODO option flags to load specific views
	// TODO read config
	// TODO get db location from config
	// TODO log rotation
	conn, err := sql.Open("sqlite", "data.sql")
	if err != nil {
		fmt.Printf("Failed to open sqlite in %s\n", "data.sql")
		os.Exit(1)
	}

	if err := goose.Setup(conn, "schema"); err != nil {
		fmt.Printf("Failed to run migrations:\n%s\n", err)
		os.Exit(1)
	}

	config := newConfig(db.New(conn))
	model := model{
		newProjectList(config),
		config,
	}

	//p := tea.NewProgram(model, tea.WithAltScreen())
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
