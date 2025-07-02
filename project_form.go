package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)


type projectData struct {
	name string
	description *string
	parent_id *int64
	planned_for *time.Time
	start_at *time.Time
	due_at *time.Time
	completed_at *time.Time
}

type projectForm struct {
	id int64
	form *huh.Form
	config *config
	initial projectData
	parents *[]formIdSelect
	width int
	height int
}

func fetchTopProjects(queries *db.Queries) tea.Cmd {
	return func () tea.Msg {
		rows, err := queries.TopProjects(context.Background())

		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		if err != nil {
			return errMsg{err}
		}

		items := make([]formIdSelect, len(rows))
		for i, e := range rows {
			items[i] = formIdSelect{e.ID, e.Name}
		}

		return formIdSelectsMsg{items}
	}
}

type createdProjectMsg int64
func (model projectForm) createProject() tea.Cmd {
	return func() tea.Msg {
		createParams := db.CreateProjectParams{Name: model.form.GetString("name")}
		if model.form.GetString("desc") != "" {
			createParams.Description = sql.NullString{Valid: true, String: model.form.GetString("desc")}
		}

		if parent, ok := model.form.Get("parent").(int64); ok && parent > 0 {
			createParams.ParentID = sql.NullInt64{Valid: true, Int64: parent}
		}

		if model.form.GetString("planned_for") != "" {
			planned, err := time.Parse("2006/01/02", model.form.GetString("planned_for"))

			if err != nil {
				return errMsg{err}
			}

			createParams.PlannedFor = sql.NullTime{Valid: true, Time: planned}
		}

		if model.form.GetString("start_at") != "" {
			start, err := time.Parse("2006/01/02", model.form.GetString("start_at"))

			if err != nil {
				return errMsg{err}
			}

			createParams.StartAt = sql.NullTime{Valid: true, Time: start}
		}

		if model.form.GetString("due_at") != "" {
			due, err := time.Parse("2006/01/02", model.form.GetString("due_at"))

			if err != nil {
				return errMsg{err}
			}

			createParams.DueAt = sql.NullTime{Valid: true, Time: due}
		}

		if model.form.GetString("completed_at") != "" {
			complete, err := time.Parse("2006/01/02", model.form.GetString("completed_at"))

			if err != nil {
				return errMsg{err}
			}

			createParams.CompletedAt = sql.NullTime{Valid: true, Time: complete}
		}

		project, err := model.config.queries.CreateProject(context.Background(), createParams)

		if err != nil {
			return errMsg{err}
		}

		// This is technically not necessary for how it's used as of writing this, however it is nice to have a distinct message should it be useful later
		return createdProjectMsg(project.ID)
	}
}

type updatedProjectMsg int64
func (model projectForm) updateProject() tea.Cmd {
	return func() tea.Msg {
		tx, err := model.config.db.Begin()
		if err != nil {
			return errMsg{err}
		}
		defer tx.Rollback()

		qtx := model.config.queries.WithTx(tx)

		if model.form.GetString("name") != "" && model.form.GetString("name") != model.initial.name {
			err := qtx.UpdateProjectName(context.Background(), db.UpdateProjectNameParams{Name: model.form.GetString("name"), ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if (model.form.GetString("desc") != "" && model.initial.description == nil) ||
			(model.initial.description != nil && model.form.GetString("desc") != *model.initial.description) {
			desc := sql.NullString{}
			if model.form.GetString("desc") != "" {
				desc.Valid = true
				desc.String = model.form.GetString("desc")
			}

			err := qtx.UpdateProjectDescription(context.Background(), db.UpdateProjectDescriptionParams{Description: desc, ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if (model.form.GetString("due_at") != "" && model.initial.due_at == nil) ||
			(model.initial.due_at != nil && model.form.GetString("due_at") != model.initial.due_at.Format("2006/01/02")) {
			due := sql.NullTime{}
			if model.form.GetString("due_at") != "" {
				due.Valid = true
				due_at_time, err := time.Parse("2006/01/02", model.form.GetString("due_at"))
				if err != nil {
					return errMsg{err}
				}
				due.Time = due_at_time
			}

			err := qtx.UpdateProjectDueAt(context.Background(), db.UpdateProjectDueAtParams{DueAt: due, ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if (model.form.GetString("planned_for") != "" && model.initial.planned_for == nil) ||
			(model.initial.planned_for != nil && model.form.GetString("planned_for") != model.initial.planned_for.Format("2006/01/02")) {
			plan := sql.NullTime{}
			if model.form.GetString("planned_for") != "" {
				plan.Valid = true
				plan_time, err := time.Parse("2006/01/02", model.form.GetString("planned_for"))
				if err != nil {
					return errMsg{err}
				}
				plan.Time = plan_time
			}

			err := qtx.UpdateProjectPlannedFor(context.Background(), db.UpdateProjectPlannedForParams{PlannedFor: plan, ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if (model.form.GetString("start_at") != "" && model.initial.start_at == nil) ||
			(model.initial.start_at != nil && model.form.GetString("start_at") != model.initial.start_at.Format("2006/01/02")) {
			start := sql.NullTime{}
			if model.form.GetString("start_at") != "" {
				start.Valid = true
				start_at_time, err := time.Parse("2006/01/02", model.form.GetString("start_at"))
				if err != nil {
					return errMsg{err}
				}
				start.Time = start_at_time
			}

			err := qtx.UpdateProjectStartAt(context.Background(), db.UpdateProjectStartAtParams{StartAt: start, ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if (model.form.GetString("completed_at") != "" && model.initial.completed_at == nil) ||
			(model.initial.completed_at != nil && model.form.GetString("completed_at") != model.initial.completed_at.Format("2006/01/02")) {
			complete := sql.NullTime{}
			if model.form.GetString("completed_at") != "" {
				complete.Valid = true
				completed_at_time, err := time.Parse("2006/01/02", model.form.GetString("completed_at"))
				if err != nil {
					return errMsg{err}
				}
				complete.Time = completed_at_time
			}

			err := qtx.UpdateProjectCompletedAt(context.Background(), db.UpdateProjectCompletedAtParams{CompletedAt: complete, ID: model.id})
			if err != nil {
				return errMsg{err}
			}
		}

		if parent, ok := model.form.Get("parent").(int64); ok {
			if (model.initial.parent_id != nil && parent > 0) || (*model.initial.parent_id != parent) {
				parentParam := sql.NullInt64{}
				if parent > 0 {
					parentParam.Valid = true
					parentParam.Int64 = parent
				}

				err = qtx.UpdateProjectParent(context.Background(), db.UpdateProjectParentParams{ParentID: parentParam, ID: model.id})
				if err != nil {
					return errMsg{err}
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			return errMsg{err}
		}

		return updatedProjectMsg(model.id)
	}
}

func newCreateProjectForm(config *config, height int, width int, parent *int64) tea.Model {
	parentPlaceholder := make([]formIdSelect, 0)
	model := projectForm{config: config, width: width, height: height, parents: &parentPlaceholder}
	if parent != nil {
		model.initial = projectData{parent_id: parent}
	}
	model.form = newProjectForm(model)

	return model
}

func newEditProjectForm(config *config, current projectData, height int, width int, id int64) tea.Model {
	parentPlaceholder := make([]formIdSelect, 0)
	model := projectForm{config: config, width: width, height: height, id: id, initial: current, parents: &parentPlaceholder}
	model.form = newProjectForm(model)

	return model
}

func newProjectForm(model projectForm) *huh.Form {
	// Copy the current values, so we can later compare for UPDATE statements
	initName := model.initial.name
	var initDesc string
	var initPlanned string
	var initStart string
	var initDue string
	var initComplete string
	var initParent int64

	if model.initial.description != nil {
		initDesc = *model.initial.description
	}

	if model.initial.planned_for != nil {
		initPlanned = model.initial.planned_for.Format("2006/01/02")
	}

	if model.initial.start_at != nil {
		initStart = model.initial.start_at.Format("2006/01/02")
	}

	if model.initial.due_at != nil {
		initDue = model.initial.due_at.Format("2006/01/02")
	}

	if model.initial.completed_at != nil {
		initComplete = model.initial.completed_at.Format("2006/01/02")
	}

	if model.initial.parent_id != nil {
		initParent = *model.initial.parent_id
	}
	confirmDefault := true

	// TODO validate the dates to be valid Golang dates of format 2006/01/02, yyyy/mm/dd
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Validate(huh.ValidateNotEmpty()).
				Title("Name").
				Value(&initName),

			huh.NewText().
				Key("desc").
				Title("Description").
				Value(&initDesc),

			huh.NewInput().
				Key("planned_for").
				Title("Planned For").
				Value(&initPlanned),

			huh.NewInput().
				Key("start_at").
				Title("Start At").
				Value(&initStart),

			huh.NewInput().
				Key("due_at").
				Title("Due At").
				Value(&initDue),

			huh.NewInput().
				Key("completed_at").
				Title("Completed At").
				Value(&initComplete),

			// It would be neater if this could be hidden, but hidden is only possible with groups, and placing this in a new group always moves it to a new "page"
			huh.NewSelect[int64]().
				Key("parent").
				Title("Parent Project").
				OptionsFunc(func() []huh.Option[int64] {
					if model.parents == nil {
						unloaded := make([]huh.Option[int64], 1)
						unloaded[0] = huh.NewOption[int64]("None", 0)
						return unloaded
					}

					opts := make([]huh.Option[int64], len(*model.parents)+1)
					opts[0] = huh.NewOption[int64]("None", 0)
					for i, e := range *model.parents {
						opts[i+1] = huh.NewOption(e.title, e.id)
					}
					return opts
				}, model.parents).
				Value(&initParent),

			huh.NewConfirm().
				Key("done").
				Affirmative("confirm").
				Negative("cancel").
				Value(&confirmDefault),
		),
	).
		WithShowHelp(false).
		WithShowErrors(true).
		WithHeight(model.height).
		WithWidth(model.width)

	return form
}

func (model projectForm) Init() tea.Cmd {
	return tea.Batch(model.form.Init(), fetchTopProjects(model.config.queries))
}

func (model projectForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		newForm := model.form.WithWidth(msg.Width-4).WithHeight(msg.Height-4)
		model.form = newForm
		return model, nil

	case formIdSelectsMsg:
		parents := msg.items
		*model.parents = parents
		return model, nil

	case createdProjectMsg, updatedProjectMsg:
		model.config.enableGlobalHotkeys()
		return model, changeView(newProjectOverview(model.config))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			model.config.enableGlobalHotkeys()
			return model, changeView(newProjectOverview(model.config))
		case "ctrl+x":
			// If done without NextField, this will cause the currently active field to not be counted...
			if model.id > 0 {
				return model, tea.Sequence(model.form.NextField(), model.updateProject())
			} else {
				return model, tea.Sequence(model.form.NextField(), model.createProject())
			}
		}
	}

	var cmds []tea.Cmd

	form, cmd := model.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		model.form = f
		cmds = append(cmds, cmd)
	}

	if model.form.State == huh.StateCompleted {
		if model.form.GetBool("done") == false {
			model.config.enableGlobalHotkeys()
			return model, changeView(newProjectOverview(model.config))
		}

		if model.id > 0 {
			return model, model.updateProject()
		} else {
			return model, model.createProject()
		}
	}

	return model, tea.Batch(cmds...)
}

func (model projectForm) View() string {
	return model.form.View()
}
