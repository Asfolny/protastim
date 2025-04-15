package main

import (
	"database/sql"
	"fmt"
	"os"

	db "github.com/Asfolny/protastim/internal/database"
	goose "github.com/Asfolny/protastim/internal/sql"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type rows struct {
	data []table.Row
}

type editingDone bool

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
	case tea.WindowSizeMsg:
		m.config.size = msg
		return m, nil

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
	lg := m.config.lg
	border := lipgloss.NormalBorder()
	wrapperStyle := lg.NewStyle().
		BorderStyle(border).
		BorderTop(false).
		BorderLeft(true).
		BorderBottom(true).
		BorderRight(true).
		BorderForeground(lipgloss.Color("63")).
		Height(m.config.size.Height-2).
		Width(m.config.size.Width-2)

	statusStyle := lg.NewStyle().
		Foreground(lipgloss.Color("63"))

	content := wrapperStyle.Render(m.view.View())
	title := statusStyle.Render(border.TopLeft + border.Top + border.Top) + " Protatastim"
	titleWidth := lipgloss.Width(title)
	status := lipgloss.PlaceHorizontal(m.config.size.Width-titleWidth, lipgloss.Right, "doing nothing " + statusStyle.Render(border.Top + border.Top + border.TopRight))


	return title + status + "\n" + content
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

	const q = `
    PRAGMA foreign_keys = ON;
    PRAGMA journal_mode = WAL;
    `
	_, err = conn.Exec(q)
	if err != nil {
		fmt.Printf("Failed to set up sqlite pragmas:\n%s\n", err)
		os.Exit(1)
	}

	if err := goose.Setup(conn, "schema"); err != nil {
		fmt.Printf("Failed to run migrations:\n%s\n", err)
		os.Exit(1)
	}

	config := newConfig(db.New(conn))
	model := model{
		view: newProjectList(config),
		config: config,
	}
	//model := model{
	//	view: newTaskList(config),
	//	config: config,
	//}

	p := tea.NewProgram(model, tea.WithAltScreen())
	//p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
