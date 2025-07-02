package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"os"

	db "github.com/Asfolny/protastim/internal/database"
	goose "github.com/Asfolny/protastim/internal/sql"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	view   tea.Model
	config *config
	tracking bool
	entry db.WorkingOnWithNameRow
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.view.Init(), func() tea.Msg {
		current, err := m.config.queries.WorkingOnWithName(context.Background())

		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		if err != nil {
			return errMsg{err}
		}

		return current
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.config.size = msg
		m.view, cmd = m.view.Update(msg)
		return m, cmd

	case startedTimerMsg:
		m.tracking = true
		m.entry = msg
		return m, nil

	case stoppedTimerMsg:
		m.tracking = false
		m.entry = db.WorkingOnWithNameRow{}
		return m, nil

	case newViewMsg:
		m.view = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "ctrl+t":
			if m.entry.TaskID > 0 {
				return m, toggleTimer(m.config.queries, m.entry.TaskID)
			}
		case "D":
			if _, ok := m.view.(dashboard); !ok {
				return m, changeView(newDashboard(m.config))
			}
		case "P":
			if _, ok := m.view.(projectOverview); !ok {
				return m, changeView(newProjectOverview(m.config))
			}
		}
	}

	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

func (m model) View() string {
	border := lipgloss.NormalBorder()
	wrapperStyle := lipgloss.NewStyle().
		BorderStyle(border).
		BorderTop(false).
		BorderLeft(true).
		BorderBottom(true).
		BorderRight(true).
		BorderForeground(lipgloss.Color("63"))

	w, h := wrapperStyle.GetFrameSize()
	titleGutter := 1
	wrapperStyle = wrapperStyle.Height(m.config.size.Height - h - titleGutter).Width(m.config.size.Width - w)

	statusStyle := lipgloss.NewStyle().Height(titleGutter).Foreground(lipgloss.Color("63"))
	title := statusStyle.Height(titleGutter).Render(border.TopLeft + border.Top + border.Top) + " Protastim"
	titleWidth := lipgloss.Width(title)

	var progress string
	if m.tracking {
		progress = fmt.Sprintf("Working on: %s - %s", m.entry.Name, time.Since(m.entry.StartAt).Round(time.Second).String())
	}

	status := lipgloss.PlaceHorizontal(m.config.size.Width-titleWidth, lipgloss.Right, progress + " " + statusStyle.Render(border.Top + border.Top + border.TopRight))
	content := wrapperStyle.Render(m.view.View())
	return fmt.Sprintf("%s%s\n%s", title, status, content)
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

	config := newConfig(conn)
	model := model{
		view: newDashboard(config),
		config: config,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
