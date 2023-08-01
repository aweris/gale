package core

// Status is the phase of the lifecycle that object currently in
type Status string

const (
	StatusQueued     Status = "queued"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// Conclusion is outcome of the operation
type Conclusion string

const (
	ConclusionSuccess   Conclusion = "success"
	ConclusionFailure   Conclusion = "failure"
	ConclusionCancelled Conclusion = "cancelled"
	ConclusionSkipped   Conclusion = "skipped"
)

//TODO: add support for docker and composite steps types

// StepType is the type of the step
type StepType string

const (
	StepTypeAction  StepType = "action"
	StepTypeRun     StepType = "run"
	StepTypeUnknown StepType = "unknown"
)
