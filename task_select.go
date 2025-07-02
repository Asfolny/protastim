package main

import (
	"context"
	"fmt"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func fetchScheduledTasks(queries *db.Queries, source string) tea.Cmd {
	return func () tea.Msg {
		row, err := queries.ScheduledTasks(context.Background())

		if err != nil {
			return errMsg{err}
		}

		items := make([]list.Item, len(row))
		for i, e := range row {
			var desc strings.Builder

			desc.WriteString(fmt.Sprintf("Project: %s", e.ProjectName))

			if e.DueAt.Valid {
				desc.WriteString(fmt.Sprintf("\nDue: %s", e.DueAt.Time.Format("2006/02/01")))
			}

			items[i] = listItem{
				title: e.Name,
				desc: desc.String(),
				id: e.ID,
				started: e.StartAt.Valid,
			}
		}

		return selectorItemsMsg{source, items}
	}
}

type startedTaskMsg int64
func startTask(queries *db.Queries, id int64) tea.Cmd {
	return func () tea.Msg {
		err := queries.StartTask(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return startedTaskMsg(id)
	}
}

type completedTaskMsg int64
func completeTask(queries *db.Queries, id int64) tea.Cmd {
	return func() tea.Msg {
		err := queries.CompleteTask(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return completedTaskMsg(id)
	}
}

func newPlannedTaskSelector(config *config, width int, height int, source string) selector {
	model := selector{config: config, width: width, height: height, fetchFunc: fetchScheduledTasks(config.queries, source), tracking: source}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	delegate.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		var id int64
		var started bool

		if i, ok := list.SelectedItem().(listItem); ok {
			id = i.id
			started = i.started
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "S":
				cmds := []tea.Cmd{
					list.NewStatusMessage(statusStyle.Render("Starting")),
					toggleTimer(config.queries, id),
				}

				if !started {
					cmds = append(cmds, startTask(config.queries, id))
				}

				return tea.Batch(cmds...)

			case "alt+s":
				cmds := []tea.Cmd{
					list.NewStatusMessage(statusStyle.Render("Starting without tracking")),
				}

				if !started {
					cmds = append(cmds, startTask(config.queries, id))
				}

				return tea.Batch(cmds...)

			case "T":
				return tea.Batch(toggleTimer(config.queries, id), list.NewStatusMessage(statusStyle.Render("Toggling timer")))

			case "C":
				return tea.Batch(list.NewStatusMessage(statusStyle.Render("Completing")), completeTask(config.queries, id))
			}
		}

		return nil
	}

	selectorList := list.New(make([]list.Item, 0), delegate, 0, 0)
	selectorList.Title = "Schedule"
	selectorList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	selectorList.Styles.NoItems = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#626262"))
	selectorList.Styles.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	selectorList.SetShowHelp(false)
	selectorList.SetSpinner(spinner.MiniDot)
	model.list = selectorList

	return model
}
