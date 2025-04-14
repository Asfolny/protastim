package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectListModel struct {
	config  *config
	table   table.Model
	err     error
	loading bool
	spinner spinner.Model
}

func newProjectList(config *config) projectListModel {
	m := projectListModel{
		config:  config,
		loading: true,
		spinner: config.newSpinner(),
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 2},
			{Title: "Status", Width: 8},
			{Title: "Name"},
		}),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(ts)

	m.table = t
	return m
}

func (m projectListModel) fetchProjects() tea.Msg {
	m.config.mu.RLock()
	defer m.config.mu.RUnlock()

	row, err := m.config.queries.ListProjects(context.Background())
	if err != nil {
		return errMsg{err}
	}

	tableRows := make([]table.Row, len(row), len(row))
	for i, e := range row {
		tableRows[i] = table.Row{strconv.FormatInt(e.ID, 10), e.Status, e.Name}
	}

	return rows{tableRows}
}

func (m projectListModel) Init() tea.Cmd {
	return tea.Batch(m.fetchProjects, m.spinner.Tick)
}

func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case rows:
		m.table.SetRows(msg.data)
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n":
			return m, changeView(newProjectEdit(m.config, nil))

		case "enter":
			id, _ := strconv.ParseInt(m.table.SelectedRow()[0], 10, 64)
			return m, changeView(newProject(m.config, id))
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m projectListModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n%s Loading\n\n", m.spinner.View())
	}

	if m.err != nil {
		return fmt.Sprintf("\nFailed fetching project: %v\n\n", m.err)
	}
	// TODO this specific handling should be done within Update instead
	cols := m.table.Columns()
	for i, col := range cols {
		if col.Title == "Name" {
			col.Width = m.config.getInnerWidth() - 4 - 8 - 2
			cols[i] = col
		}
	}
	m.table.SetColumns(cols)
	m.table.SetHeight(m.config.getInnerHeight())
	m.table.SetWidth(m.config.size.Width)

	lg := m.config.lg

	return lg.NewStyle().PaddingTop(1).Render(m.table.View())
}
