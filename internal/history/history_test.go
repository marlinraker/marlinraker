package history

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"gotest.tools/v3/assert"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/files"
)

func setup(t *testing.T) {
	t.Helper()
	notification.Testing = true
	files.Fs = afero.NewMemMapFs()
	files.DataDir = "/data"
	assert.NilError(t, afero.WriteFile(files.Fs, "/data/gcodes/benchy.gcode", []byte("G28\n"), 0644))
	assert.NilError(t, database.Init())
	assert.NilError(t, Init())
}

func TestJobLifecycle(t *testing.T) {
	setup(t)

	id1 := JobStarted("benchy.gcode")
	id2 := JobStarted("missing.gcode")
	assert.Equal(t, id1, "000001")
	assert.Equal(t, id2, "000002")

	JobFinished(id1, StatusCompleted, 90*time.Second, 100*time.Second, 1234.5)
	JobFinished(id2, StatusCancelled, 10*time.Second, 20*time.Second, -1)

	job, err := GetJob(id1)
	assert.NilError(t, err)
	assert.Equal(t, job.Status, StatusCompleted)
	assert.Equal(t, job.Exists, true)
	assert.Equal(t, job.PrintDuration, 90.)
	assert.Equal(t, job.TotalDuration, 100.)
	assert.Equal(t, job.FilamentUsed, 1234.5)
	assert.Assert(t, job.EndTime != nil)

	job, err = GetJob(id2)
	assert.NilError(t, err)
	assert.Equal(t, job.Exists, false)
	assert.Equal(t, job.FilamentUsed, 0.)

	_, err = GetJob("FFFFFF")
	assert.ErrorContains(t, err, "invalid job uid")

	assert.DeepEqual(t, GetTotals(), Totals{
		TotalJobs:         2,
		TotalTime:         120,
		TotalPrintTime:    100,
		TotalFilamentUsed: 1234.5,
		LongestJob:        100,
		LongestPrint:      90,
	})

	// newest first by default
	jobs, count := List(ListOptions{Limit: 1})
	assert.Equal(t, count, 2)
	assert.Equal(t, len(jobs), 1)
	assert.Equal(t, jobs[0].JobID, id2)

	jobs, _ = List(ListOptions{Asc: true})
	assert.Equal(t, jobs[0].JobID, id1)

	jobs, _ = List(ListOptions{Start: 5})
	assert.Equal(t, len(jobs), 0)

	// state survives a restart and running jobs become interrupted
	id3 := JobStarted("benchy.gcode")
	assert.NilError(t, Init())
	job, err = GetJob(id3)
	assert.NilError(t, err)
	assert.Equal(t, job.Status, StatusInterrupted)
	assert.Equal(t, GetTotals().TotalJobs, 2)
	assert.Equal(t, JobStarted("benchy.gcode"), "000004")

	deleted, err := DeleteJob(id1, false)
	assert.NilError(t, err)
	assert.DeepEqual(t, deleted, []string{id1})
	_, err = DeleteJob(id1, false)
	assert.ErrorContains(t, err, "invalid job uid")

	deleted, err = DeleteJob("", true)
	assert.NilError(t, err)
	assert.Equal(t, len(deleted), 3)
	_, count = List(ListOptions{})
	assert.Equal(t, count, 0)

	// deleting jobs keeps the totals
	last, err := ResetTotals()
	assert.NilError(t, err)
	assert.Equal(t, last.TotalJobs, 2)
	assert.DeepEqual(t, GetTotals(), Totals{})
}
