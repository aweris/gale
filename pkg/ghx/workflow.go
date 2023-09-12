package ghx

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/idgen"
)

var _ TaskResult = new(WorkflowResult)

// WorkflowResult represents the result of a workflow
type WorkflowResult struct {
	RunID      string                `json:"run_id"`     // RunID is the run id of the job run
	Conclusion core.Conclusion       `json:"conclusion"` // Conclusion is the result of a completed step after continue-on-error is applied
	Jobs       map[string]*JobResult `json:"jobs"`       // Jobs is map of the job run id to its result
}

// GetConclusion returns the conclusion of the task.
func (r WorkflowResult) GetConclusion() core.Conclusion {
	return r.Conclusion
}

// WorkflowRunner is the runner that executes the workflow.
type WorkflowRunner struct {
	RunID          string                       // RunID is the ID of the run
	RunNumber      string                       // RunNumber is the number of the run
	RunAttempt     string                       // RunAttempt is the attempt number of the run
	RetentionDays  string                       // RetentionDays is the number of days to keep the run logs
	Workflow       core.Workflow                // Workflow is the workflow to run
	ExecutionOrder []string                     // ExecutionOrder is the order of the jobs to run
	Jobs           map[string]*JobResult        // Jobs is map of the job run id to its result
	executor       TaskExecutor[WorkflowResult] // executor is the main task executor that executes the jobs and keeps the execution information.
}

// Plan plans the workflow and returns the workflow runner.
func Plan(wf core.Workflow, job string) (*WorkflowRunner, error) {
	runID, err := idgen.GenerateWorkflowRunID()
	if err != nil {
		return nil, err
	}

	runner := &WorkflowRunner{
		RunID:         runID,
		RunNumber:     "1",
		RunAttempt:    "1",
		RetentionDays: "0",
		Workflow:      wf,
		Jobs:          make(map[string]*JobResult),
	}

	jrs := make([]*JobRunner, 0)

	for _, jm := range wf.Jobs {
		if jm.ID != job {
			jobRunID, err := idgen.GenerateJobRunID()
			if err != nil {
				return nil, err
			}

			runner.Jobs[jm.ID] = &JobResult{
				RunID:      jobRunID,
				Conclusion: core.ConclusionSkipped,
				Outcome:    core.ConclusionSkipped,
				Outputs:    make(map[string]string),
			}

			continue
		}

		jr, err := planJob(jm)
		if err != nil {
			return nil, err
		}

		jrs = append(jrs, jr)

		runner.Jobs[jm.ID] = &JobResult{Outputs: make(map[string]string)}
	}

	if len(jrs) == 0 {
		return nil, fmt.Errorf("job %s not found in workflow %s", job, wf.Name)
	}

	runner.executor = NewTaskExecutor[WorkflowResult](fmt.Sprintf("Workflow: %s", wf.Name), func(ctx context.Context) (*WorkflowResult, error) {
		for _, jr := range jrs {
			_, jobResult, err := jr.Run(ctx)

			runner.Jobs[jr.Job.ID] = jobResult

			if err != nil {
				res := &WorkflowResult{
					RunID:      runID,
					Conclusion: core.ConclusionFailure,
					Jobs:       runner.Jobs,
				}

				return res, err
			}
		}

		result := &WorkflowResult{RunID: runner.RunID, Conclusion: core.ConclusionSuccess, Jobs: runner.Jobs}

		return result, nil
	})

	return runner, nil
}

// Run runs the workflow..
func (r *WorkflowRunner) Run(ctx context.Context) error {
	// ignore the result for now, we will use it later
	_, _, err := r.executor.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
