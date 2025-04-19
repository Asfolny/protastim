package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
)

type dashboardFocus int
const (
	dashboardSelectorFocus dashboardFocus = iota
	dashboardTaskDetailsFocus
)

type dashboard struct {
	config   *config
	selector tea.Model
	taskDetails tea.Model
	focus  dashboardFocus
	err error
}

func newDashboard(c *config) tea.Model {
	return dashboard{config: c, selector: newScheduledTaskSelector(c, c.getInnerWidth() / 5 * 2, c.getInnerHeight()-1)}
}

func (model dashboard) Init() tea.Cmd {
	return model.selector.Init()
}

func (model dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		model.selector, cmd = model.selector.Update(tea.WindowSizeMsg{Width: model.config.getInnerWidth() / 5 * 2, Height: model.config.getInnerHeight()-1})
		cmds = append(cmds, cmd)

		if model.taskDetails != nil {
			model.taskDetails, cmd = model.taskDetails.Update(tea.WindowSizeMsg{Width: model.config.getInnerWidth() / 5 * 3, Height: model.config.getInnerHeight()-1})
			cmds = append(cmds, cmd)
		}

		return model, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		model.selector, cmd = model.selector.Update(msg)
		cmds = append(cmds, cmd)

		if model.taskDetails != nil {
			model.taskDetails, cmd = model.taskDetails.Update(msg)
			cmds = append(cmds, cmd)
		}

		return model, tea.Batch(cmds...)

	case selectedItemMsg:
		model.taskDetails = newTaskView(model.config, msg, model.config.getInnerWidth() / 5 * 3 , model.config.getInnerHeight() - 1)
		return model, model.taskDetails.Init()

	case selectorItemsMsg:
		var cmd tea.Cmd
		model.selector, cmd = model.selector.Update(msg)
		return model, cmd

	case fetchedTask:
		var cmd tea.Cmd
		model.taskDetails, cmd = model.taskDetails.Update(msg)
		return model, cmd

	case errMsg:
		model.err = msg.err
		return model, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			switch model.focus {
			case dashboardSelectorFocus:
				model.focus = dashboardTaskDetailsFocus
				return model, nil

			case dashboardTaskDetailsFocus:
				model.focus = dashboardSelectorFocus
				return model, nil
			}
		}
	}

	switch model.focus {
	case dashboardSelectorFocus:
		var cmd tea.Cmd
		model.selector, cmd = model.selector.Update(msg)
		return model, cmd

	case dashboardTaskDetailsFocus:
		var cmd tea.Cmd
		model.taskDetails, cmd = model.taskDetails.Update(msg)
		return model, cmd
	}

	return model, nil
}

func (model dashboard) View() string {
	dashboardStyle := lipgloss.NewStyle().
		MarginTop(1)

	selectorView := model.selector.View()
	taskDetailsView := lipgloss.NewStyle(). // Setting a default style, as unselected it may be nil
		BorderStyle(lipgloss.ASCIIBorder()).
		BorderBackground(lipgloss.Color("99")).
		Width(model.config.getInnerWidth() / 5 * 3).
		Height(model.config.getInnerHeight() - 1).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Nothing selected")

	if model.taskDetails != nil {
		taskDetailsView = model.taskDetails.View()
	}

	if model.err != nil {
		return dashboardStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, selectorView, model.err.Error()))
	}

	return dashboardStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, selectorView, taskDetailsView))
}
