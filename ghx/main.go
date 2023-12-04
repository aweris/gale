package main

import (
	stdContext "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
)

func main() {
	stdctx := stdContext.Background()

	client, err := dagger.Connect(stdctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fmt.Printf("failed to get dagger client: %v", err)
		os.Exit(1)
	}

	// Load context
	ctx, err := context.New(stdctx, client)
	if err != nil {
		fmt.Printf("failed to load context: %v", err)
		os.Exit(1)
	}

	cfg := ctx.GhxConfig

	// Load workflow
	wf, err := LoadWorkflow(cfg, filepath.Join(cfg.HomeDir, "run", "workflow.yaml"))
	if err != nil {
		fmt.Printf("could not load workflow: %v", err)
		os.Exit(1)
	}

	ctx.Execution.Workflow = &wf

	report := &context.WorkflowRunReport{
		Ran:           false,
		Name:          wf.Name,
		Path:          wf.Path,
		RunID:         ctx.Github.RunID,
		RunNumber:     ctx.Github.RunNumber,
		RunAttempt:    ctx.Github.RunAttempt,
		RetentionDays: ctx.Github.RetentionDays,
		Conclusion:    model.ConclusionSuccess,
		Jobs:          make(map[string]model.Conclusion),
	}

	// Check if the report file exists
	if _, err := os.Stat(filepath.Join(cfg.HomeDir, "run", "workflow_run.json")); err == nil {
		// The file exists, read it
		data, err := os.ReadFile(filepath.Join(cfg.HomeDir, "run", "workflow_run.json"))
		if err != nil {
			fmt.Printf("failed to read workflow run report: %v", err)
			os.Exit(1)
		}

		// Unmarshal the data into the report structure
		if err := json.Unmarshal(data, report); err != nil {
			fmt.Printf("failed to unmarshal workflow run report: %v", err)
			os.Exit(1)
		}
	}

	jm, ok := wf.Jobs[ctx.GhxConfig.Job]
	if !ok {
		fmt.Printf("job %s not found", ctx.GhxConfig.Job)
		os.Exit(1)
	}

	runners, err := planJob(jm)
	if err != nil {
		fmt.Printf("failed to plan job: %v", err)
		os.Exit(1)
	}

	// FIXME: ignoring fail-fast for now. it is always true for now. Fix this later.
	// FIXME: run all runners sequentially for now. Ignoring parallelism. Fix this later.

	for _, runner := range runners {
		result, err := runner.Run(ctx)
		if err != nil {
			fmt.Printf("failed to run job: %v", err)
			os.Exit(1)
		}

		if report.Conclusion == model.ConclusionSuccess && result.Conclusion != report.Conclusion {
			report.Conclusion = result.Conclusion
		}
	}

	report.Ran = true
	report.Duration = "n/a"

	data, err := json.Marshal(report)
	if err != nil {
		fmt.Printf("failed to marshal workflow run report: %v", err)
		os.Exit(1)
	}

	err = os.WriteFile(filepath.Join(cfg.HomeDir, "run", "workflow_run.json"), data, 0644)
	if err != nil {
		fmt.Printf("failed to write workflow run report: %v", err)
		os.Exit(1)
	}
}

func LoadWorkflow(cfg context.GhxConfig, path string) (model.Workflow, error) {
	var workflow model.Workflow

	if err := fs.ReadYAMLFile(path, &workflow); err != nil {
		return workflow, err
	}

	// set workflow path
	workflow.Path = cfg.Workflow

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = cfg.Workflow
	}

	// update job ID and names
	for idj, job := range workflow.Jobs {
		job.ID = idj

		if job.Name == "" {
			job.Name = idj
		}

		// update step IDs if not provided
		for ids, step := range job.Steps {
			if step.ID == "" {
				step.ID = fmt.Sprintf("%d", ids)
			}

			job.Steps[ids] = step
		}

		workflow.Jobs[idj] = job
	}

	return workflow, nil
}
