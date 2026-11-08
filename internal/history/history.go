// Package history records print jobs, compatible with Moonraker's history component.
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/api/notification"
	"github.com/marlinraker/marlinraker/internal/database"
	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/util"
)

// Job statuses as used by Moonraker
const (
	StatusInProgress  = "in_progress"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
	StatusError       = "error"
	StatusInterrupted = "interrupted"
)

// maxJobs limits the number of stored jobs; the oldest jobs are dropped first.
const maxJobs = 1000

type Job struct {
	JobID         string         `json:"job_id"`
	Exists        bool           `json:"exists"`
	EndTime       *float64       `json:"end_time"`
	FilamentUsed  float64        `json:"filament_used"`
	Filename      string         `json:"filename"`
	Metadata      map[string]any `json:"metadata"`
	PrintDuration float64        `json:"print_duration"`
	Status        string         `json:"status"`
	StartTime     float64        `json:"start_time"`
	TotalDuration float64        `json:"total_duration"`
}

type Totals struct {
	TotalJobs         int     `json:"total_jobs"`
	TotalTime         float64 `json:"total_time"`
	TotalPrintTime    float64 `json:"total_print_time"`
	TotalFilamentUsed float64 `json:"total_filament_used"`
	LongestJob        float64 `json:"longest_job"`
	LongestPrint      float64 `json:"longest_print"`
}

// state is persisted as a single database item.
type state struct {
	NextID int    `json:"next_id"`
	Jobs   []Job  `json:"jobs"`
	Totals Totals `json:"totals"`
}

const (
	dbNamespace = "history"
	dbKey       = "state"
)

var (
	mu      sync.Mutex
	current state
)

// Init loads the history from the database. Jobs that were still in progress
// when Marlinraker stopped are marked as interrupted.
func Init() error {
	mu.Lock()
	defer mu.Unlock()

	current = state{NextID: 1}
	value, err := database.GetItem(dbNamespace, dbKey, true)
	if err != nil {
		var executorError *util.ExecutorError
		if errors.As(err, &executorError) && executorError.Code == 404 {
			return nil
		}
		return fmt.Errorf("failed to load history: %w", err)
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}
	if err := json.Unmarshal(bytes, &current); err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	changed := false
	for i := range current.Jobs {
		if current.Jobs[i].Status == StatusInProgress {
			current.Jobs[i].Status = StatusInterrupted
			changed = true
		}
	}
	if changed {
		return save()
	}
	return nil
}

func save() error {
	if _, err := database.PostItem(dbNamespace, dbKey, current, true); err != nil {
		return fmt.Errorf("failed to save history: %w", err)
	}
	return nil
}

func now() float64 {
	return float64(time.Now().UnixMilli()) / 1000
}

func publish(action string, job Job) {
	job.Exists = fileExists(job.Filename)
	err := notification.Publish(notification.New("notify_history_changed", []any{map[string]any{
		"action": action,
		"job":    job,
	}}))
	if err != nil {
		log.Errorf("Failed to publish history change: %v", err)
	}
}

func fileExists(fileName string) bool {
	_, err := files.Fs.Stat(filepath.Join(files.DataDir, "gcodes", fileName))
	return err == nil
}

func loadMetadata(fileName string) map[string]any {
	metadata, err := files.LoadOrScanMetadata(fileName)
	if err != nil || metadata == nil {
		return map[string]any{}
	}
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return map[string]any{}
	}
	result := map[string]any{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return map[string]any{}
	}
	return result
}

// JobStarted records a new print job and returns its id.
func JobStarted(fileName string) string {
	metadata := loadMetadata(fileName)

	mu.Lock()
	defer mu.Unlock()

	job := Job{
		JobID:     fmt.Sprintf("%06X", current.NextID),
		Filename:  fileName,
		Metadata:  metadata,
		Status:    StatusInProgress,
		StartTime: now(),
	}
	current.NextID++
	current.Jobs = append(current.Jobs, job)
	if len(current.Jobs) > maxJobs {
		current.Jobs = slices.Delete(current.Jobs, 0, len(current.Jobs)-maxJobs)
	}

	if err := save(); err != nil {
		log.Error(err)
	}
	publish("added", job)
	return job.JobID
}

// JobFinished updates a job when it ends and adds it to the totals.
func JobFinished(jobID string, status string, printDuration, totalDuration time.Duration, filamentUsed float64) {
	mu.Lock()
	defer mu.Unlock()

	idx := slices.IndexFunc(current.Jobs, func(job Job) bool { return job.JobID == jobID })
	if idx == -1 {
		log.Warnf("Finished job %s is not in the history", jobID)
		return
	}

	job := &current.Jobs[idx]
	end := now()
	job.EndTime = &end
	job.Status = status
	job.PrintDuration = printDuration.Seconds()
	job.TotalDuration = totalDuration.Seconds()
	job.FilamentUsed = math.Max(0, filamentUsed)

	totals := &current.Totals
	totals.TotalJobs++
	totals.TotalTime += job.TotalDuration
	totals.TotalPrintTime += job.PrintDuration
	totals.TotalFilamentUsed += job.FilamentUsed
	totals.LongestJob = math.Max(totals.LongestJob, job.TotalDuration)
	totals.LongestPrint = math.Max(totals.LongestPrint, job.PrintDuration)

	if err := save(); err != nil {
		log.Error(err)
	}
	publish("finished", *job)
}

type ListOptions struct {
	Start  int
	Limit  int
	Since  float64
	Before float64
	Asc    bool
}

// List returns the matching jobs and the number of all matching jobs.
func List(options ListOptions) ([]Job, int) {
	mu.Lock()
	jobs := slices.Clone(current.Jobs)
	mu.Unlock()

	jobs = slices.DeleteFunc(jobs, func(job Job) bool {
		return (options.Since > 0 && job.StartTime <= options.Since) ||
			(options.Before > 0 && job.StartTime >= options.Before)
	})
	if !options.Asc {
		slices.Reverse(jobs)
	}

	count := len(jobs)
	start := min(max(options.Start, 0), count)
	end := count
	if options.Limit > 0 {
		end = min(start+options.Limit, count)
	}

	result := jobs[start:end]
	for i := range result {
		result[i].Exists = fileExists(result[i].Filename)
	}
	return result, count
}

// GetJob returns the job with the given id.
func GetJob(jobID string) (Job, error) {
	mu.Lock()
	defer mu.Unlock()
	idx := slices.IndexFunc(current.Jobs, func(job Job) bool { return job.JobID == jobID })
	if idx == -1 {
		return Job{}, util.NewErrorf(404, "invalid job uid %q", jobID)
	}
	job := current.Jobs[idx]
	job.Exists = fileExists(job.Filename)
	return job, nil
}

// DeleteJob deletes the job with the given id or all jobs if all is true.
// It returns the ids of the deleted jobs.
func DeleteJob(jobID string, all bool) ([]string, error) {
	mu.Lock()
	defer mu.Unlock()

	var deleted []string
	if all {
		for _, job := range current.Jobs {
			deleted = append(deleted, job.JobID)
		}
		current.Jobs = nil
	} else {
		idx := slices.IndexFunc(current.Jobs, func(job Job) bool { return job.JobID == jobID })
		if idx == -1 {
			return nil, util.NewErrorf(404, "invalid job uid %q", jobID)
		}
		deleted = []string{jobID}
		current.Jobs = slices.Delete(current.Jobs, idx, idx+1)
	}

	if deleted == nil {
		deleted = []string{}
	}
	return deleted, save()
}

// GetTotals returns the accumulated job totals.
func GetTotals() Totals {
	mu.Lock()
	defer mu.Unlock()
	return current.Totals
}

// ResetTotals resets the totals and returns the previous ones.
func ResetTotals() (Totals, error) {
	mu.Lock()
	defer mu.Unlock()
	last := current.Totals
	current.Totals = Totals{}
	return last, save()
}
