package models_test

import (
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink/adapters"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/utils"
	"github.com/stretchr/testify/assert"
	null "gopkg.in/guregu/null.v3"
)

func TestJobSave(t *testing.T) {
	t.Parallel()
	store, cleanup := cltest.NewStore()
	defer cleanup()

	j1 := cltest.NewJobWithSchedule("* * * * 7")
	assert.Nil(t, store.SaveJob(j1))

	store.Save(j1)
	j2, _ := store.FindJob(j1.ID)
	assert.Equal(t, j1.Initiators[0].Schedule, j2.Initiators[0].Schedule)
}

func TestJobNewRun(t *testing.T) {
	t.Parallel()

	job := cltest.NewJobWithSchedule("1 * * * *")
	job.Tasks = []models.Task{models.Task{Type: "NoOp"}}

	newRun := job.NewRun()
	assert.Equal(t, job.ID, newRun.JobID)
	assert.Equal(t, 1, len(newRun.TaskRuns))
	assert.Equal(t, "NoOp", job.Tasks[0].Type)
	assert.Nil(t, job.Tasks[0].Params)
	adapter, _ := adapters.For(job.Tasks[0])
	assert.NotNil(t, adapter)
}

func TestJobEnded(t *testing.T) {
	t.Parallel()

	endAt := utils.ParseNullableTime("3000-01-01T00:00:00.000Z")

	tests := []struct {
		name    string
		endAt   null.Time
		current time.Time
		want    bool
	}{
		{"no end at", null.Time{Valid: false}, endAt.Time, false},
		{"before end at", endAt, endAt.Time.Add(-time.Nanosecond), false},
		{"at end at", endAt, endAt.Time, false},
		{"after end at", endAt, endAt.Time.Add(time.Nanosecond), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			job := cltest.NewJob()
			job.EndAt = test.endAt

			assert.Equal(t, test.want, job.Ended(test.current))
		})
	}
}

func TestJobStarted(t *testing.T) {
	t.Parallel()

	startAt := utils.ParseNullableTime("3000-01-01T00:00:00.000Z")

	tests := []struct {
		name    string
		startAt null.Time
		current time.Time
		want    bool
	}{
		{"no start at", null.Time{Valid: false}, startAt.Time, true},
		{"before start at", startAt, startAt.Time.Add(-time.Nanosecond), false},
		{"at start at", startAt, startAt.Time, true},
		{"after start at", startAt, startAt.Time.Add(time.Nanosecond), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			job := cltest.NewJob()
			job.StartAt = test.startAt

			assert.Equal(t, test.want, job.Started(test.current))
		})
	}
}
