package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
)

type listItem struct {
	title string
	desc string
	id int64
	started bool
	completed bool
}

func (item listItem) Title() string {
	return item.title
}

func (item listItem) Description() string {
	return item.desc
}

func (item listItem) FilterValue() string {
	return item.title
}

type selectorItemsMsg struct {
	source string
	items []list.Item
}

type selectedItemMsg = int64
func (model selector) changeItem(item list.Item) tea.Cmd {
	if i, ok := item.(listItem); ok {
		return func() tea.Msg {
			return selectedItemMsg(i.id)
		}
	}

	return nil
}

type selector struct {
	config *config
	list   list.Model
	fetchFunc tea.Cmd
	tracking string
	err    error
	width  int
	height int
	ignoreChange bool
}

func (model selector) Init() tea.Cmd {
	return tea.Batch(model.fetchFunc, model.list.StartSpinner())
}

func (model selector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.list.SetSize(model.width, model.height)
		return model, nil

	case selectorItemsMsg:
		if model.tracking == msg.source {
			model.list.SetItems(msg.items)
			model.list.StopSpinner()
			i, ok := model.list.SelectedItem().(listItem)


			if len(msg.items) > 0 && ok && model.ignoreChange == false {
				return model, func() tea.Msg {
					return selectedItemMsg(i.id)
				}
			}
		}
		return model, nil

	case errMsg:
		model.err = msg.err
		model.list.StopSpinner()
		return model, nil

	case startedTaskMsg, completedTaskMsg:
		return model, tea.Batch(model.fetchFunc, model.list.StartSpinner())

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			return model, model.fetchFunc
		}
	}

	var cmds []tea.Cmd
	prevItem := model.list.SelectedItem()

	newList, cmd := model.list.Update(msg)
	cmds = append(cmds, cmd)
	model.list = newList

	if model.ignoreChange == false && model.list.SelectedItem() != nil && model.list.FilterState() != list.Filtering && prevItem != model.list.SelectedItem() {
		cmds = append(cmds, model.changeItem(model.list.SelectedItem()))
	}

	return model, tea.Batch(cmds...)
}

func (model selector) View() string {
	if model.err != nil {
		return lipgloss.PlaceHorizontal(model.width, lipgloss.Center, model.err.Error())
	}

	if len(model.list.Items()) < 1 {
		model.list.SetShowStatusBar(false)
	}

	return lipgloss.NewStyle().Height(model.height).Width(model.width).Render(model.list.View())
}
