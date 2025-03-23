package main

import (
	"context"
	"fmt"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type taskModel struct {
	id         int64
	config     *config
	data       *db.Task
	modal      tea.Model
	err        error
	loading    bool
	editing    bool
	spinner    spinner.Model
}

func newTask(config *config, id int64) taskModel {
	m := taskModel{
		id:      id,
		config: config,
		loading: true,
		spinner: config.newSpinner(),
	}

	return m
}

func (m taskModel) fetchTask() tea.Msg {
	row, err := m.config.queries.GetTask(context.Background(), m.id)
	if err != nil {
		return errMsg{err}
	}

	return row
}

func (m taskModel) Init() tea.Cmd {
	return tea.Batch(m.fetchTask, m.spinner.Tick)
}

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case db.Task:
		m.data = &msg
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n":
			m.modal = newTaskEdit(m.config, nil)
			m.editing = true
			return m, m.modal.Init()
		}
	}

	if m.editing {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m taskModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n%s Loading\n\n", m.spinner.View())
	}

	if m.err != nil {
		return fmt.Sprintf("\nFailed fetching task: %v\n\n", m.err)
	}

	if m.editing {
		return m.modal.View()
	}

	return fmt.Sprintf(
		"Name: %s\nDescription:\n%s\nStatus: %s\n",
		m.data.Name,
		m.data.Description.String,
		m.data.Status,
	)
}
