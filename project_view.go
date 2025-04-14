package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectViewFocusElement int

const (
	focusDesc projectViewFocusElement = iota
	focusTasks
)

type projectModel struct {
	id             int64
	config         *config
	data           *db.Project
	err            error
	loadingProject bool
	loadingTasks   bool
	spinner        spinner.Model
	descVw         viewport.Model
	taskList       table.Model
	focus          projectViewFocusElement
}

func newProject(config *config, id int64) projectModel {
	m := projectModel{
		id:             id,
		config:         config,
		loadingProject: true,
		loadingTasks:   true,
		spinner:        config.newSpinner(),
		descVw:         viewport.New(0, 0),
		focus:          focusTasks,
	}

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 4},
			{Title: "Name"},
			{Title: "Status", Width: 8},
		}),
		table.WithFocused(true),
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

	m.taskList = t

	return m
}

func (m projectModel) fetchProject() tea.Msg {
	project, err := m.config.queries.GetProject(context.Background(), m.id)
	if err != nil {
		return errMsg{err}
	}

	return project
}

func (m projectModel) fetchTasks() tea.Msg {
	tasks, err := m.config.queries.ListTasksByProject(context.Background(), m.id)
	if err != nil {
		return errMsg{err}
	}

	tableRows := make([]table.Row, len(tasks), len(tasks))
	for i, e := range tasks {
		tableRows[i] = table.Row{strconv.FormatInt(e.ID, 10), e.Name, e.Status}
	}

	return rows{tableRows}
}

func (m projectModel) Init() tea.Cmd {
	return tea.Batch(m.fetchProject, m.fetchTasks, m.spinner.Tick)
}

func (m projectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case rows:
		m.loadingTasks = false
		m.taskList.SetRows(msg.data)
		return m, nil

	case db.Project:
		m.data = &msg
		m.loadingProject = false
		m.descVw.SetContent(m.data.Description.String)

		return m, nil

	case errMsg:
		m.err = msg
		m.loadingProject = false
		m.loadingTasks = false
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			switch m.focus {
			case focusDesc:
				m.focus = focusTasks
				m.taskList.Focus()
			case focusTasks:
				m.focus = focusDesc
				m.taskList.Blur()
			}
		case "enter":
			if m.focus == focusTasks {
				id, _ := strconv.ParseInt(m.taskList.SelectedRow()[0], 10, 64)
				return m, changeView(newTask(m.config, id))
			}
		case "ctrl+n":
			return m, changeView(newProjectEdit(m.config, nil))
		case "ctrl+e":
			return m, changeView(newProjectEdit(m.config, m.data))
		case "ctrl+t":
			return m, changeView(newTaskEdit(m.config, nil, m.data.Name))
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case focusDesc:
		m.descVw, cmd = m.descVw.Update(msg)
	case focusTasks:
		m.taskList, cmd = m.taskList.Update(msg)
	}

	return m, cmd
}

func (m projectModel) View() string {
	if m.loadingProject {
		return fmt.Sprintf("\n%s Loading\n\n", m.spinner.View())
	}

	if m.err != nil {
		return fmt.Sprintf("\nFailed fetching project or tasks: %v\n\n", m.err)
	}

	var sb strings.Builder
	lg := m.config.lg
	titleStyle := lg.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(m.config.getInnerWidth()).
		MarginTop(2).
		MarginBottom(1).
		Bold(true)

	title := titleStyle.Render(m.data.Name + "\n")
	status := m.config.styles.Base.Render(fmt.Sprintf("Status: %s\n", m.data.Status))
	// TODO more task/desc layouts
	desc := m.descVw
	desc.SetContent(m.data.Description.String)
	desc.Width = (m.config.getInnerWidth() / 2) - 4
	desc.Height = m.config.getInnerHeight() - lipgloss.Height(title) - lipgloss.Height(status)

	cols := m.taskList.Columns()
	for i, col := range cols {
		if col.Title == "Name" {
			col.Width = desc.Width - 4 - 8
			cols[i] = col
		}
	}
	m.taskList.SetColumns(cols)
	m.taskList.SetHeight(desc.Height)
	m.taskList.SetWidth(desc.Width)

	descOut := lg.NewStyle().BorderStyle(lipgloss.NormalBorder()).Render(desc.View())
	table := lg.NewStyle().BorderStyle(lipgloss.NormalBorder()).Render(m.taskList.View())
	sb.WriteString(title)
	sb.WriteString(status)
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, descOut, table))

	return sb.String()
}
