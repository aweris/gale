package model

import "time"

// WorkflowRunStatus represents the status of a workflow run. It can be one of the following:
// completed, action_required, cancelled, failure, neutral, skipped, stale, success, timed_out, in_progress,
// queued, requested, waiting, pending
//
// For now, we only care about queued, completed, in_progress, skipped, success, failure.
type WorkflowRunStatus string

const (
	WorkflowRunStatusQueued     WorkflowRunStatus = "queued"
	WorkflowRunStatusCompleted  WorkflowRunStatus = "completed"
	WorkflowRunStatusInProgress WorkflowRunStatus = "in_progress"
	WorkflowRunStatusSkipped    WorkflowRunStatus = "skipped"
	WorkflowRunStatusSuccess    WorkflowRunStatus = "success"
	WorkflowRunStatusFailure    WorkflowRunStatus = "failure"
)

// WorkflowRun represents a workflow run.
type WorkflowRun struct {
	ID           string            `json:"id"`             // ID is the unique identifier for the workflow run.
	Name         string            `json:"name"`           // Name is the name of the workflow run.
	Status       WorkflowRunStatus `json:"status"`         // Status is the current status of the workflow run.
	RunStartedAt time.Time         `json:"run_started_at"` // RunStartedAt is the time the workflow run started.
	RunDuration  time.Duration     `json:"run_duration"`   // RunDuration is the duration of the workflow run.
}
