package context

import (
	"time"

	"github.com/aweris/gale/common/model"
)

// RunResult is the result of the task execution. The signature of the RunResult is identical to task.Result to avoid
// circular dependency. // FIXME: decouple this from task.Result completely
type RunResult struct {
	Ran        bool             `json:"ran"`        // Ran indicates if the execution ran
	Conclusion model.Conclusion `json:"conclusion"` // Conclusion of the execution
	Duration   time.Duration    `json:"duration"`   // Duration of the execution
}

type WorkflowRunReport struct {
	Ran           bool                        `json:"ran"`            // Ran indicates if the execution ran
	Duration      string                      `json:"duration"`       // Duration of the execution
	Name          string                      `json:"name"`           // Name is the name of the workflow
	Path          string                      `json:"path"`           // Path is the path of the workflow
	RunID         string                      `json:"run_id"`         // RunID is the ID of the run
	RunNumber     string                      `json:"run_number"`     // RunNumber is the number of the run
	RunAttempt    string                      `json:"run_attempt"`    // RunAttempt is the attempt number of the run
	RetentionDays string                      `json:"retention_days"` // RetentionDays is the number of days to keep the run logs
	Conclusion    model.Conclusion            `json:"conclusion"`     // Conclusion is the result of a completed workflow run after continue-on-error is applied
	Jobs          map[string]model.Conclusion `json:"jobs"`           // Jobs is map of the job run id to its result
}

type JobRunReport struct {
	Ran        bool                    `json:"ran"`               // Ran indicates if the execution ran
	Duration   string                  `json:"duration"`          // Duration of the execution
	Name       string                  `json:"name"`              // Name is the name of the job
	RunID      string                  `json:"run_id"`            // RunID is the ID of the run
	Conclusion model.Conclusion        `json:"conclusion"`        // Conclusion is the result of a completed job after continue-on-error is applied
	Outcome    model.Conclusion        `json:"outcome"`           // Outcome is  the result of a completed job before continue-on-error is applied
	Outputs    map[string]string       `json:"outputs,omitempty"` // Outputs is the outputs generated by the job
	Matrix     model.MatrixCombination `json:"matrix,omitempty"`  // Matrix is the matrix parameters used to run the job
	Steps      []StepRunSummary        `json:"steps"`             // Steps is the list of steps in the job
}

type StepRunSummary struct {
	ID         string           `json:"id"`             // ID is the unique identifier of the step.
	Name       string           `json:"name,omitempty"` // Name is the name of the step
	Stage      model.StepStage  `json:"stage"`          // Stage is the stage of the step during the execution of the job. Possible values are: setup, pre, main, post, complete.
	Conclusion model.Conclusion `json:"conclusion"`     // Conclusion is the result of a completed job after continue-on-error is applied
}

// NewJobRunReport creates a new job run report from the given job run.
func NewJobRunReport(result *RunResult, jr *model.JobRun) *JobRunReport {
	report := &JobRunReport{
		Ran:        result.Ran,
		Duration:   result.Duration.String(),
		Conclusion: result.Conclusion,
		Name:       jr.Job.Name,
		RunID:      jr.RunID,
		Outcome:    jr.Outcome,
		Outputs:    jr.Outputs,
		Matrix:     jr.Matrix,
	}

	for _, step := range jr.Steps {
		summary := StepRunSummary{
			ID:         step.Step.ID,
			Name:       step.Step.Name,
			Stage:      step.Stage,
			Conclusion: step.Conclusion,
		}

		report.Steps = append(report.Steps, summary)
	}

	return report
}

type StepRunReport struct {
	Ran        bool              `json:"ran"`               // Ran indicates if the execution ran
	Duration   string            `json:"duration"`          // Duration of the execution
	ID         string            `json:"id"`                // ID is the unique identifier of the step.
	Name       string            `json:"name,omitempty"`    // Name is the name of the step
	Conclusion model.Conclusion  `json:"conclusion"`        // Conclusion is the result of a completed job after continue-on-error is applied
	Outcome    model.Conclusion  `json:"outcome"`           // Outcome is  the result of a completed job before continue-on-error is applied
	Outputs    map[string]string `json:"outputs,omitempty"` // Outputs is the outputs generated by the job
	State      map[string]string `json:"state,omitempty"`   // State is a map of step state variables.
	Env        map[string]string `json:"env,omitempty"`     // Env is the extra environment variables set by the step.
	Path       []string          `json:"path,omitempty"`    // Path is extra PATH items set by the step.
}

// NewStepRunReport creates a new step run report from the given step run.
func NewStepRunReport(result *RunResult, sr *model.StepRun) *StepRunReport {
	return &StepRunReport{
		Ran:        result.Ran,
		Duration:   result.Duration.String(),
		ID:         sr.Step.ID,
		Name:       sr.Step.Name,
		Conclusion: result.Conclusion,
		Outcome:    sr.Outcome,
		Outputs:    sr.Outputs,
		State:      sr.State,
		Env:        sr.Environment,
		Path:       sr.Path,
	}
}
