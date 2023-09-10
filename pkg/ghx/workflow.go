package ghx

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/idgen"
)

// WorkflowRunner is the runner that executes the workflow.
type WorkflowRunner struct {
	RunID          string                    // RunID is the ID of the run
	RunNumber      string                    // RunNumber is the number of the run
	RunAttempt     string                    // RunAttempt is the attempt number of the run
	RetentionDays  string                    // RetentionDays is the number of days to keep the run logs
	Workflow       core.Workflow             // Workflow is the workflow to run
	ExecutionOrder []string                  // ExecutionOrder is the order of the jobs to run
	Jobs           map[string]core.JobResult // Jobs is map of the job run id to its result
	executor       TaskExecutor              // executor is the main task executor that executes the jobs and keeps the execution information.
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
		Jobs:          make(map[string]core.JobResult),
	}

	var jrs []*JobRunner

	for _, jm := range wf.Jobs {
		if jm.ID != job {
			jobRunID, err := idgen.GenerateJobRunID()
			if err != nil {
				return nil, err
			}

			runner.Jobs[jm.ID] = core.JobResult{
				ID:         jm.ID,
				Name:       jm.Name,
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

		runner.Jobs[jm.ID] = core.JobResult{
			ID:      jm.ID,
			Name:    jm.Name,
			RunID:   jr.RunID,
			Outputs: make(map[string]string),
		}
	}

	if len(jrs) == 0 {
		return nil, fmt.Errorf("job %s not found in workflow %s", job, wf.Name)
	}

	runner.executor = NewTaskExecutor(fmt.Sprintf("Workflow: %s", wf.Name), func(ctx context.Context) (conclusion core.Conclusion, err error) {
		for _, jr := range jrs {
			result := runner.Jobs[jr.Job.ID]

			if err := jr.Run(ctx); err != nil {
				result.Conclusion = core.ConclusionFailure
				result.Outcome = core.ConclusionFailure

				runner.Jobs[jr.Job.ID] = result

				return core.ConclusionFailure, err
			}

			result.Conclusion = core.ConclusionSuccess
			result.Outcome = core.ConclusionSuccess

			runner.Jobs[jr.Job.ID] = result
		}

		return core.ConclusionSuccess, nil
	})

	return runner, nil
}

// Run runs the workflow..
func (r *WorkflowRunner) Run(ctx context.Context) error {
	// run is always true for the main task executor and concussion not important.
	_, _, err := r.executor.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
