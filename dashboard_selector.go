package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
)

type listItem struct {
	title string
	desc string
	id int64
	started bool
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

type selectorItemsMsg = []list.Item

func (selector dashboardSelector) fetchScheduledTasks() tea.Msg {
	row, err := selector.config.queries.ScheduledTasks(context.Background())

	if err != nil {
		return errMsg{err}
	}

	items := make([]list.Item, len(row))
	for i, e := range row {
		var desc strings.Builder

		desc.WriteString(fmt.Sprintf("Project: %s", e.ProjectName))

		if e.PlannedFor.Valid {
			desc.WriteString(fmt.Sprintf("\nScheduled: %s", e.PlannedFor.Time.Format("2006/02/01")))
		}

		if e.DueAt.Valid {
			desc.WriteString(fmt.Sprintf("\nDue: %s", e.DueAt.Time.Format("2006/02/01")))
		}

		if e.StartAt.Valid {
			desc.WriteString(fmt.Sprintf("\nStarted: %s", e.StartAt.Time.Format("2006/02/01")))
		}

		items[i] = listItem{
			title: e.Name,
			desc: desc.String(),
			id: e.ID,
			started: e.StartAt.Valid,
		}
	}

	return selectorItemsMsg(items)
}


type startedTimerMsg = db.TimeEntry
type stoppedTimerMsg any // always nil, just a type for msg

func (selector dashboardSelector) toggleTimer(id int64) tea.Cmd {
	return func() tea.Msg {
		timerStarted := true
		currentTimer, err := selector.config.queries.GetRunningTimer(context.Background())

		if errors.Is(err, sql.ErrNoRows) {
			timerStarted = false
			err = nil
		}

		if err != nil {
			return errMsg{err}
		}

		if timerStarted {
			err := selector.config.queries.StopTimeTracking(context.Background(), sql.NullTime{Time: time.Now(), Valid: true})
			if err != nil {
				return errMsg{err}
			}

			if currentTimer.TaskID == id {
				return stoppedTimerMsg(nil)
			}
		}

		timeEntry, err := selector.config.queries.StartTimeTracking(context.Background(), db.StartTimeTrackingParams{TaskID: id, StartAt: time.Now()} )
		if err != nil {
			return errMsg{err}
		}

		return startedTimerMsg(timeEntry)
	}
}

type startedTaskMsg struct {
	taskId int64
}

func (selector dashboardSelector) startTask(id int64) tea.Cmd {
	return func () tea.Msg {
		err := selector.config.queries.StartTask(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return startedTaskMsg{id}
	}
}

type selectedItemMsg = int64

func (selector dashboardSelector) changeItem(item list.Item) tea.Cmd {
	if i, ok := item.(listItem); ok {
		return func() tea.Msg {
			return selectedItemMsg(i.id)
		}
	}

	return nil
}

type dashboardSelector struct {
	config *config
	list   list.Model
	err    error
	width  int
	height int
}

func newDashboardSelector(c *config, width int, height int) dashboardSelector {
	model := dashboardSelector{config: c, width: width, height: height}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(5)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	delegate.UpdateFunc = func(msg tea.Msg, list *list.Model) tea.Cmd {
		var title string
		var id int64
		var started bool

		if i, ok := list.SelectedItem().(listItem); ok {
			title = i.title
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
					list.NewStatusMessage(statusStyle.Render("Starting " + title)),
					model.toggleTimer(id),
				}

				if !started {
					cmds = append(cmds, model.startTask(id))
				}

				return tea.Batch(cmds...)
			case "alt+s":
				cmds := []tea.Cmd{
					list.NewStatusMessage(statusStyle.Render("Starting " + title)),
				}

				if !started {
					cmds = append(cmds, model.startTask(id))
				}

				return tea.Batch(cmds...)
			}
		}

		return nil
	}

	selectorList := list.New(make([]list.Item, 0), delegate, 0, 0)
	selectorList.Title = "Plans"
	selectorList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	selectorList.Styles.NoItems = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#626262"))
	selectorList.Styles.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	selectorList.SetShowHelp(false)
	selectorList.SetSpinner(spinner.MiniDot)
	model.list = selectorList

	return model
}

func (model dashboardSelector) Init() tea.Cmd {
	return tea.Batch(model.fetchScheduledTasks, model.list.StartSpinner())
}

func (model dashboardSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.list.SetSize(model.width, model.height)
		return model, nil

	case selectorItemsMsg:
		model.list.SetItems(msg)
		model.list.StopSpinner()
		if len(msg) > 0 {
			if i, ok := msg[0].(listItem); ok {
				return model, func () tea.Msg {
					return selectedItemMsg(i.id)
				}
			}
		}
		return model, nil

	case errMsg:
		model.err = msg.err
		model.list.StopSpinner()
		return model, nil

	case startedTaskMsg:
		return model, tea.Batch(model.fetchScheduledTasks, model.list.StartSpinner())
	}

	var cmds []tea.Cmd
	prevItem := model.list.SelectedItem()

	newList, cmd := model.list.Update(msg)
	cmds = append(cmds, cmd)
	model.list = newList

	if model.list.SelectedItem() != nil && model.list.FilterState() != list.Filtering && prevItem != model.list.SelectedItem() {
		cmds = append(cmds, model.changeItem(model.list.SelectedItem()))
	}

	return model, tea.Batch(cmds...)
}

func (model dashboardSelector) View() string {
	if model.err != nil {
		return lipgloss.PlaceHorizontal(model.width, lipgloss.Center, model.err.Error())
	}

	if len(model.list.Items()) < 1 {
		model.list.SetShowStatusBar(false)
	}

	return lipgloss.NewStyle().Height(model.height).Width(model.width).Render(model.list.View())
}
