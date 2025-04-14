package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type taskEditModel struct {
	config *config
	form   *huh.Form
	width  int
	create bool
	loading *bool
	projects *[]db.Project
	project string
}

func newTaskEdit(config *config, task *db.Task, project string) tea.Model {
	t := true
	var p []db.Project
	m := taskEditModel{config: config, create: task == nil, loading: &t, projects: &p, project: project}
	confirmDefault := true

	var name string
	var desc string
	var status string

	if task != nil {
		name = task.Name
		desc = task.Description.String
		status = task.Status
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Name").
				Prompt("> ").
				Validate(huh.ValidateNotEmpty()).
				Value(&name),

			huh.NewText().
				Key("description").
				Title("Description").
				Value(&desc),

			huh.NewSelect[string]().
				Key("status").
				Options(huh.NewOptions("N", "P", "C", "H", "D")...).
				Title("Status ").
				Inline(true).
				Value(&status),

			huh.NewSelect[string]().
				Key("project").
				Title("Project").
				OptionsFunc(func() []huh.Option[string] {
					if *m.loading {
						empty := make([]string, 2)
						empty[0] = "Loading..."
						empty[1] = m.project
						return huh.NewOptions(empty...)
					}

					return huh.NewOptions(m.projectsAsOptions()...)
				}, m.loading).
				Validate(func(val string) error {
					if (val == "Choose a project") {
						return errors.New("Must choose a project for task")
					}

					return nil
				}).
				Value(&m.project),

			huh.NewConfirm().
				Validate(func(v bool) error { // Should SHOULD close
					if !v {
						return fmt.Errorf("Welp, finish up then")
					}
					return nil
				}).
				Affirmative("done").
				Negative("cancel").
				Inline(true).
				Value(&confirmDefault),
		),
	).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

	return m
}

func (m taskEditModel) projectsAsOptions() []string {
	l := make([]string, len(*m.projects) + 1)
	l[0] = "Choose a project"
	for i, e := range *m.projects {
		l[i+1] = e.Name
		i++
	}

	return l
}

func (m taskEditModel) fetchProjects() tea.Msg {
	row, err := m.config.queries.ListProjects(context.Background())
	if err != nil {
		return errMsg{err}
	}

	return row
}

func (m taskEditModel) getProjectId() int64 {
	for _, e := range *m.projects {
		if e.Name == m.form.GetString("project") {
			return e.ID
		}
	}

	return 0
}


func (m taskEditModel) Init() tea.Cmd {
	return tea.Batch(m.form.Init(), m.fetchProjects)
}

func (m taskEditModel) createTaskCmd() func() tea.Msg {
	m.config.mu.Lock()
	defer m.config.mu.Unlock()

	desc := sql.NullString{String: m.form.GetString("description")}
	if desc.String != "" {
		desc.Valid = true
	}


	dataset := db.CreateTaskParams{
		Name:        m.form.GetString("name"),
		Description: desc,
		Status:      m.form.GetString("status"),
		ProjectID:   m.getProjectId(),
	}

	return func() tea.Msg {
		m.config.queries.CreateTask(context.Background(), dataset)
		return editingDone(true)
	}
}

func (m taskEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.config.styles.Base.GetHorizontalFrameSize()

	case []db.Project:
		*m.projects = msg
		*m.loading = false
		return m, nil

	case editingDone:
		return m, changeView(newProject(m.config, m.getProjectId()))
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// TODO handle abort
		// TODO handle updating an existing
		cmds = append(cmds, m.createTaskCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m taskEditModel) View() string {
	var status string
	s := m.config.styles
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.config.lg.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()
	var header string

	if m.create {
		header = m.appBoundaryView("Create Task")
	} else {
		header = m.appBoundaryView("Update Task")
	}

	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m taskEditModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m taskEditModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.config.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m taskEditModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.config.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
