package main

import (
	"context"
	"fmt"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


type taskView struct {
	config *config
	taskId int64
	task   db.Task
	loading bool
	spinner spinner.Model
	desc viewport.Model
	height int
	width int
}

type fetchedTask = db.Task

func (model taskView) fetchTaskData() tea.Msg {
	task, err:= model.config.queries.GetTask(context.Background(), model.taskId)
	if err != nil {
		return errMsg{err}
	}

	return task
}

func newTaskView(config *config, id int64, width int, height int) taskView {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	vp := viewport.New(width, 0)
	vp.Style = vp.Style.Padding(1, 1)
	return taskView{config: config, taskId: id, loading: true, spinner: s, desc: vp, height: height, width: width}
}

func (model taskView) Init() tea.Cmd {
	return tea.Batch(model.spinner.Tick, model.fetchTaskData)
}

func (model taskView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.desc.Width = msg.Width
		return model, nil

	case fetchedTask:
		model.task = msg
		model.loading = false

		if msg.Description.Valid {
			model.desc.SetContent(msg.Description.String)
		}

		return model, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)
		return model, cmd
	}

	return model, nil
}

func (model taskView) View() string {
	style := lipgloss.NewStyle().Width(model.width).Height(model.height)

	if model.loading {
		return style.Align(lipgloss.Center, lipgloss.Center).Render(model.spinner.View())
	}

	var sb strings.Builder

	nameStyle := lipgloss.NewStyle().Width(model.width).AlignHorizontal(lipgloss.Center)
	sb.WriteString(nameStyle.Render(model.task.Name) + "\n")

	if model.task.CompletedAt.Valid {
		sb.WriteString(fmt.Sprintf("\nCompleted: %s", model.task.CompletedAt.Time.Format("2006/02/01")))
	}

	if model.task.DueAt.Valid {
		sb.WriteString(fmt.Sprintf("\nDue: %s", model.task.DueAt.Time.Format("2006/02/01")))
	}

	if model.task.PlannedFor.Valid {
		sb.WriteString(fmt.Sprintf("\nScheduled: %s", model.task.PlannedFor.Time.Format("2006/02/01")))
	}

	if model.task.StartAt.Valid {
		sb.WriteString(fmt.Sprintf("\nStarted: %s", model.task.StartAt.Time.Format("2006/02/01")))
	}

	model.desc.Height = model.height - lipgloss.Height(sb.String())

	return style.Render(fmt.Sprintf("%s\n%s", sb.String(), model.desc.View()))
}
