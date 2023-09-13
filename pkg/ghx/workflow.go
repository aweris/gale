package ghx

import (
	"errors"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/idgen"
)

var ErrWorkflowNotFound = errors.New("workflow not found")

// WorkflowRunner is the runner that executes the workflow.
type WorkflowRunner struct {
	context        *gctx.Context          // context is the expression context for the workflow run.
	RunID          string                 // RunID is the ID of the run
	RunNumber      string                 // RunNumber is the number of the run
	RunAttempt     string                 // RunAttempt is the attempt number of the run
	RetentionDays  string                 // RetentionDays is the number of days to keep the run logs
	Workflow       core.Workflow          // Workflow is the workflow to run
	ExecutionOrder []string               // ExecutionOrder is the order of the jobs to run
	Jobs           map[string]core.JobRun // Jobs is map of the job run id to its result
	taskRunner     TaskRunner             // taskRunner is the main task taskRunner that executes the jobs and keeps the execution information.
}

// Plan plans the workflow and returns the workflow runner.
func Plan(rc *gctx.Context, workflow, job string) (*WorkflowRunner, error) {
	wf, ok := rc.Repo.Workflows[workflow]
	if !ok {
		return nil, ErrWorkflowNotFound
	}

	runID, err := idgen.GenerateWorkflowRunID()
	if err != nil {
		return nil, err
	}

	runner := &WorkflowRunner{
		context:       rc,
		RunID:         runID,
		RunNumber:     "1",
		RunAttempt:    "1",
		RetentionDays: "0",
		Workflow:      wf,
		Jobs:          make(map[string]core.JobRun),
	}

	jrs := make([]*JobRunner, 0)

	for _, jm := range wf.Jobs {
		if jm.ID != job {
			jobRunID, err := idgen.GenerateJobRunID()
			if err != nil {
				return nil, err
			}

			runner.Jobs[jm.ID] = core.JobRun{
				RunID:      jobRunID,
				Conclusion: core.ConclusionSkipped,
				Outcome:    core.ConclusionSkipped,
				Outputs:    make(map[string]string),
			}

			continue
		}

		jr, err := planJob(rc, jm)
		if err != nil {
			return nil, err
		}

		jrs = append(jrs, jr)

		runner.Jobs[jm.ID] = core.JobRun{Outputs: make(map[string]string)}
	}

	if len(jrs) == 0 {
		return nil, fmt.Errorf("job %s not found in workflow %s", job, wf.Name)
	}

	runFn := func(ctx *gctx.Context) (core.Conclusion, error) {
		for _, jr := range jrs {
			_, conclusion, err := jr.Run(ctx)

			result, ok := runner.Jobs[jr.Job.ID]
			if !ok {
				result = core.JobRun{RunID: jr.RunID, Outputs: make(map[string]string)}
			}

			result.Conclusion = conclusion
			result.Outcome = conclusion

			runner.Jobs[jr.Job.ID] = result

			if err != nil {
				return core.ConclusionFailure, err
			}
		}

		return core.ConclusionSuccess, nil
	}

	opt := TaskOpts{
		ConditionalFn: nil,
		PreRunFn:      newTaskPreRunFnForWorkflow(runID, wf),
		PostRunFn:     newTaskPostRunFnForWorkflow(),
	}
	runner.taskRunner = NewTaskRunner(fmt.Sprintf("Workflow: %s", wf.Name), runFn, opt)

	return runner, nil
}

// Run runs the workflow..
func (r *WorkflowRunner) Run() error {
	// ignore the result for now, we will use it later
	_, _, err := r.taskRunner.Run(r.context)
	if err != nil {
		return err
	}

	return nil
}

func newTaskPreRunFnForWorkflow(runID string, wf core.Workflow) TaskPreRunFn {
	return func(ctx *gctx.Context) error {

		ctx.SetWorkflow(
			&core.WorkflowRun{
				RunID:         runID,
				RunNumber:     "1",
				RunAttempt:    "1",
				RetentionDays: "0",
				Workflow:      wf,
			},
		)

		return nil
	}
}

func newTaskPostRunFnForWorkflow() TaskPostRunFn {
	return func(ctx *gctx.Context) (err error) {

		ctx.UnsetWorkflow()

		return nil
	}
}
