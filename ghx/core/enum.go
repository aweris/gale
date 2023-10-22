package core

// Conclusion is outcome of the operation
type Conclusion string

const (
	ConclusionSuccess   Conclusion = "success"
	ConclusionFailure   Conclusion = "failure"
	ConclusionCancelled Conclusion = "cancelled"
	ConclusionSkipped   Conclusion = "skipped"
)

// TODO: add support for docker and composite steps types

// StepType is the type of the step
type StepType string

const (
	// StepTypeAction represents a step uses a custom action to run.
	//
	// See: https://docs.github.com/en/actions/creating-actions/about-actions#types-of-actions
	StepTypeAction StepType = "action"

	// StepTypeDocker represents a step uses a docker image to run.
	//
	// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#example-using-a-docker-hub-action
	StepTypeDocker StepType = "docker"

	// StepTypeRun represents a step uses a shell command to run.
	//
	// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun
	StepTypeRun StepType = "run"

	// StepTypeUnknown represents a step type that is not supported by the runner.
	StepTypeUnknown StepType = "unknown"
)

// StepStage is the stage of the step during the execution of the job. Possible values are: setup, pre, main, post, complete.
type StepStage string

const (
	StepStagePre  StepStage = "pre"
	StepStageMain StepStage = "main"
	StepStagePost StepStage = "post"
)
