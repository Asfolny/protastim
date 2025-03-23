package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type projectEditModel struct {
	config *config
	form   *huh.Form
	width  int
	create bool
}

func newProjectEdit(config *config, project *db.Project) tea.Model {
	m := projectEditModel{config: config, create: project == nil}
	confirmDefault := true

	var name string
	var desc string
	var status string

	if project != nil {
		name = project.Name
		desc = project.Description.String
		status = project.Status
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

			// TODO make a select between "create and open" and "only create"
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

func (m projectEditModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m projectEditModel) createProjectCmd() func() tea.Msg {
	m.config.mu.Lock()
	defer m.config.mu.Unlock()

	desc := sql.NullString{String: m.form.GetString("description")}
	if desc.String != "" {
		desc.Valid = true
	}

	dataset := db.CreateProjectParams{
		Name:        m.form.GetString("name"),
		Description: desc,
		Status:      m.form.GetString("status"),
	}

	return func() tea.Msg {
		m.config.queries.CreateProject(context.Background(), dataset)
		return created(true)
	}
}

func (m projectEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// TODO make custom func for tea.Cmd and set it on m.form.SubmitCmd to run the db's createProjectCmd function
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.config.styles.Base.GetHorizontalFrameSize()

	case created:
		return m, changeView(newProjectList(m.config))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
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
		cmds = append(cmds, m.createProjectCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m projectEditModel) View() string {
	var status string
	s := m.config.styles
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.config.lg.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()
	header := m.appBoundaryView("Charm Employment Application")
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

func (m projectEditModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m projectEditModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.config.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m projectEditModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.config.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
