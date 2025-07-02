package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectFocus int
const (
	projectSelectorFocus projectFocus = iota
	projectDetailsFocus
)

type projectOverview struct {
	config *config
	selector tea.Model
	details tea.Model
	focus projectFocus
	err error
}

func newProjectOverview(c *config) tea.Model {
	return projectOverview{config: c, selector: newProjectSelector(c, c.getInnerWidth() / 5 * 2, c.getInnerHeight()-2, "projects")}
}

func (model projectOverview) Init() tea.Cmd {
	return model.selector.Init()
}

func (model projectOverview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		model.selector, cmd = model.selector.Update(tea.WindowSizeMsg{Width: model.config.getInnerWidth() / 5 * 2, Height: model.config.getInnerHeight()-2})
		cmds = append(cmds, cmd)

		if model.details != nil {
			model.details, cmd = model.details.Update(tea.WindowSizeMsg{Width: model.config.getInnerWidth() / 5 * 3, Height: model.config.getInnerHeight()-2})
			cmds = append(cmds, cmd)
		}

		return model, tea.Batch(cmds...)

	case spinner.TickMsg, selectorItemsMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		model.selector, cmd = model.selector.Update(msg)
		cmds = append(cmds, cmd)

		if model.details != nil {
			model.details, cmd = model.details.Update(msg)
			cmds = append(cmds, cmd)
		}

		return model, tea.Batch(cmds...)

	case selectedItemMsg:
		model.details = newProjectView(model.config, msg, model.config.getInnerWidth() / 5 * 3, model.config.getInnerHeight() - 2)
		return model, model.details.Init()

	case completedProjectMsg:
		var cmd tea.Cmd
		model.selector, cmd = model.selector.Update(msg)
		return model, cmd

	case fetchedProject:
		var cmd tea.Cmd
		model.details, cmd = model.details.Update(msg)
		return model, cmd

	case errMsg:
		model.err = msg.err
		return model, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if model.focus == projectSelectorFocus {
				model.focus = projectDetailsFocus
			}
			return model, nil

		case "esc":
			if model.focus == projectDetailsFocus {
				model.focus = projectSelectorFocus
			}
			return model, nil

		case "N":
			model.config.disableGlobalHotkeys()
			return model, changeView(newCreateProjectForm(model.config, model.config.getInnerHeight(), model.config.getInnerWidth(), nil))
		}
	}

	switch model.focus {
	case projectSelectorFocus:
		var cmd tea.Cmd
		model.selector, cmd = model.selector.Update(msg)
		return model, cmd
	case projectDetailsFocus:
		var cmd tea.Cmd
		model.details, cmd = model.details.Update(msg)
		return model, cmd
	}

	return model, nil
}

func (model projectOverview) View() string {
	dashboardStyle := lipgloss.NewStyle().MarginTop(1)

	var selectorView string
	if model.focus == projectSelectorFocus {
		selectorView = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false).
			BorderForeground(lipgloss.Color("#a6da95")).
			BorderBottom(true).
			Render(model.selector.View())
	} else {
		selectorView = model.selector.View()
	}

	detailsView := lipgloss.NewStyle(). // Setting a default style, as unselected it may be nil
		Width(model.config.getInnerWidth() / 5 * 3).
		Height(model.config.getInnerHeight() - 2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Nothing selected")

	if model.details != nil {
		if model.focus == projectDetailsFocus {
			detailsView = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false).BorderForeground(lipgloss.Color("#a6da95")).BorderBottom(true).Render(model.details.View())
		} else {
			detailsView = model.details.View()
		}
	}

	if model.err != nil {
		return dashboardStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, selectorView, model.err.Error()))
	}

	return dashboardStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, selectorView, detailsView))
}
