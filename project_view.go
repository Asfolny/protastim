package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	db "github.com/Asfolny/protastim/internal/database"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loadingState int
const (
	noLoad loadingState = iota
	loadProject
	loadTasks
	loadComplete
)

const projectTaskListSource = "project-tasks"

type projectView struct {
	config *config
	projectId int64
	project db.Project
	taskList tea.Model
	loading loadingState
	spinner spinner.Model
	desc viewport.Model
	height int
	width int
}

type fetchedProject = db.Project
func fetchProjectData(q *db.Queries, id int64) tea.Cmd {
	return func() tea.Msg {
		project, err:= q.GetProject(context.Background(), id)
		if err != nil {
			return errMsg{err}
		}

		return project
	}
}

func fetchTasksByProject(queries *db.Queries, projectId int64, source string) tea.Cmd {
	return func () tea.Msg {
		row, err := queries.GetTasksByProject(context.Background(), projectId)

		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		if err != nil {
			return errMsg{err}
		}

		items := make([]list.Item, len(row))
		for i, e := range row {
			var desc strings.Builder

			if e.DueAt.Valid {
				desc.WriteString(fmt.Sprintf("Due: %s", e.DueAt.Time.Format("2006/02/01")))
			}

			if e.StartAt.Valid {
				if desc.String() != "" {
					desc.WriteString("\n")
				}
				desc.WriteString(fmt.Sprintf("Started: %s", e.StartAt.Time.Format("2006/02/01")))
			}

			if e.PlannedFor.Valid {
				if desc.String() != "" {
					desc.WriteString("\n")
				}
				desc.WriteString(fmt.Sprintf("Planned: %s", e.PlannedFor.Time.Format("2006/02/01")))
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

func newTaskList(config *config, id int64) selector {
	model := selector{config: config, fetchFunc: fetchTasksByProject(config.queries, id, projectTaskListSource), tracking: projectTaskListSource, ignoreChange: true}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)

	taskList := list.New(make([]list.Item, 0), delegate, 8, 8)
	taskList.Styles.NoItems = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#626262"))
	taskList.Styles.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	taskList.SetSpinner(spinner.MiniDot)
	taskList.SetShowHelp(false)
	taskList.SetShowStatusBar(false)
	taskList.SetShowTitle(false)
	model.list = taskList

	return model
}

func newProjectView(config *config, id int64, width int, height int) projectView {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	vp := viewport.New(width, 0)
	vp.Style = vp.Style.Padding(1, 1)
	return projectView{config: config, projectId: id, taskList: newTaskList(config, id), spinner: s, desc: vp, height: height, width: width}
}

func (model projectView) Init() tea.Cmd {
	return tea.Batch(model.spinner.Tick, fetchProjectData(model.config.queries, model.projectId), model.taskList.Init())
}

func (model projectView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		model.desc.Width = msg.Width
		return model, nil

	case fetchedProject:
		model.project = msg
		model.loading |= loadProject

		if msg.Description.Valid {
			model.desc.SetContent(msg.Description.String)
		}

		return model, nil

	case selectorItemsMsg:
		if msg.source == projectTaskListSource {
			var cmd tea.Cmd
			model.taskList, cmd = model.taskList.Update(msg)
			model.loading |= loadTasks
			return model, cmd
		}
		return model, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		model.spinner, cmd = model.spinner.Update(msg)
		return model, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "E":
			data := projectData{name: model.project.Name}
			if model.project.Description.Valid {
				desc := model.project.Description.String
				data.description = &desc
			}

			if model.project.ParentID.Valid {
				parent := model.project.ParentID.Int64
				data.parent_id = &parent
			}

			if model.project.ParentID.Valid {
				parent := model.project.ParentID.Int64
				data.parent_id = &parent
			}

			if model.project.PlannedFor.Valid {
				plan := model.project.PlannedFor.Time
				data.planned_for = &plan
			}

			if model.project.StartAt.Valid {
				start := model.project.StartAt.Time
				data.start_at = &start
			}

			if model.project.DueAt.Valid {
				due := model.project.DueAt.Time
				data.due_at = &due
			}

			if model.project.CompletedAt.Valid {
				complete := model.project.CompletedAt.Time
				data.completed_at = &complete
			}
			model.config.disableGlobalHotkeys()
			return model, changeView(newEditProjectForm(model.config, data, model.config.getInnerHeight(), model.config.getInnerWidth(), model.project.ID))
		}
	}

	return model, nil
}

func (model projectView) View() string {
	style := lipgloss.NewStyle().Width(model.width).Height(model.height)

	if model.loading != loadComplete {
		return style.Align(lipgloss.Center, lipgloss.Center).Render(model.spinner.View())
	}

	var sb strings.Builder

	nameStyle := lipgloss.NewStyle().Width(model.width).AlignHorizontal(lipgloss.Center)
	sb.WriteString(nameStyle.Render(model.project.Name) + "\n")

	if model.project.CompletedAt.Valid {
		sb.WriteString(fmt.Sprintf("\nCompleted: %s", model.project.CompletedAt.Time.Format("2006/02/01")))
	}

	if model.project.DueAt.Valid {
		sb.WriteString(fmt.Sprintf("\nDue: %s", model.project.DueAt.Time.Format("2006/02/01")))
	}

	if model.project.PlannedFor.Valid {
		sb.WriteString(fmt.Sprintf("\nScheduled: %s", model.project.PlannedFor.Time.Format("2006/02/01")))
	}

	if model.project.StartAt.Valid {
		sb.WriteString(fmt.Sprintf("\nStarted: %s", model.project.StartAt.Time.Format("2006/02/01")))
	}

	remainingHeight := (model.height - lipgloss.Height(sb.String())) / 2
	model.desc.Height = min(remainingHeight, model.desc.TotalLineCount() + model.desc.Style.GetVerticalFrameSize())

	selecter := model.taskList.(selector)
	selecter.height = remainingHeight
	selecter.width = model.width
	selecter.list.SetSize(model.width, remainingHeight)
	selecterStyle := lipgloss.NewStyle()

	sb.WriteString(lipgloss.JoinVertical(lipgloss.Top, model.desc.View(), selecterStyle.Render(selecter.View())))

	return style.Render(fmt.Sprintf("%s\n", sb.String()))
}
