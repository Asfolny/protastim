package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/Asfolny/protastim/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

type startedTimerMsg = db.WorkingOnWithNameRow
type stoppedTimerMsg time.Duration

func toggleTimer(queries *db.Queries, taskId int64) tea.Cmd {
	return func() tea.Msg {
		timerStarted := true
		currentTimer, err := queries.GetRunningTimer(context.Background())

		if errors.Is(err, sql.ErrNoRows) {
			timerStarted = false
			err = nil
		}

		if err != nil {
			return errMsg{err}
		}

		if timerStarted {
			err := queries.StopTimeTracking(context.Background(), sql.NullTime{Time: time.Now(), Valid: true})
			if err != nil {
				return errMsg{err}
			}

			if currentTimer.TaskID == taskId {
				return stoppedTimerMsg(time.Since(currentTimer.StartAt))
			}
		}

		err = queries.StartTimeTracking(context.Background(), db.StartTimeTrackingParams{TaskID: taskId, StartAt: time.Now()} )
		if err != nil {
			return errMsg{err}
		}

		timeEntry, err := queries.WorkingOnWithName(context.Background())
		if err != nil {
			return errMsg{err}
		}

		return startedTimerMsg(timeEntry)
	}
}
