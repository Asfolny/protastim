package main

import (
	"fmt"
	"context"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
)

func fetchProjects(queries *db.Queries, source string) tea.Cmd {
	return func() tea.Msg {
		rows, err := queries.ProjectsWithRecName(context.Background())
		if err != nil {
			return errMsg{err}
		}

		items := make([]list.Item, len(rows))
		for i, e := range rows {
			var desc strings.Builder

			if e.ParentTree.Valid {
				desc.WriteString(fmt.Sprintf("Parent: %s", e.ParentTree.String))
				if e.DueAt.Valid {
					desc.WriteString("\n")
				}
			}

			if e.DueAt.Valid {
				desc.WriteString(fmt.Sprintf("Due: %s", e.DueAt.Time.Format("2006/02/01")))
			}

			items[i] = listItem{
				title: e.Name,
				desc: desc.String(),
				id: e.ID,
				started: e.StartAt.Valid,
				completed: e.CompletedAt.Valid,
			}
		}

		return selectorItemsMsg{source, items}
	}
}


type startedProjectMsg int64
func startProject(queries *db.Queries, id int64) tea.Cmd {
	return func() tea.Msg {
		err := queries.StartProject(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return startedProjectMsg(id)
	}
}

type completedProjectMsg int64
func completeProject(queries *db.Queries, id int64) tea.Cmd {
	return func() tea.Msg {
		err := queries.CompleteProject(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return completedProjectMsg(id)
	}
}

func newProjectSelector(config *config, width int, height int, source string) selector {
	model := selector{config: config, width: width, height: height, fetchFunc: fetchProjects(config.queries, source), tracking: source}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	delegate.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		item, ok := list.SelectedItem().(listItem)
		if !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "S":
				if item.started {
					return nil
				}

				return tea.Batch(
					list.NewStatusMessage(statusStyle.Render("Starting " + item.title)),
					startProject(config.queries, item.id),
				)

			case "C":
				if item.completed {
					return nil
				}

				return tea.Batch(
					list.NewStatusMessage(statusStyle.Render("Completing " + item.title)),
					completeProject(config.queries, item.id),
				)
			}
		}

		return nil
	}

	selectorList := list.New(make([]list.Item, 0), delegate, width, height)
	selectorList.Title = "Projects"
	selectorList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	selectorList.Styles.NoItems = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#626262"))
	selectorList.Styles.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	selectorList.SetShowHelp(false)
	selectorList.SetSpinner(spinner.MiniDot)
	model.list = selectorList

	return model
}
