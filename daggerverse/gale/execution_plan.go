package main

import (
	"context"
	"fmt"
	"path/filepath"
)

type WorkflowExecutionPlan struct {
	// options for the workflow run.
	opts *WorkflowRunOpts

	// unique ID of the run.
	runID string

	// information about the repository.
	repo *RepoInfo

	// the workflow to run.
	workflow *Workflow
}

func NewWorkflowExecutionPlan(ctx context.Context, opts *WorkflowRunOpts) (*WorkflowExecutionPlan, error) {
	var rid = "1"

	// load repository information
	info, err := internal.RepoInfo(ctx, opts.Source, opts.Repo, opts.Branch, opts.Tag)
	if err != nil {
		return nil, err
	}

	// set workflow config
	workflow, err := internal.getWorkflow(ctx, info, opts.WorkflowFile, opts.Workflow, opts.WorkflowsDir)
	if err != nil {
		return nil, err
	}

	wep := &WorkflowExecutionPlan{
		opts:     opts,
		runID:    rid,
		repo:     info,
		workflow: workflow,
	}

	return wep, nil
}

func (wep *WorkflowExecutionPlan) Execute(ctx context.Context) (*WorkflowRun, error) {
	plans, err := wep.PlanJobs()
	if err != nil {
		return nil, err
	}

	runs := make([]*JobRun, 0, len(plans))

	for _, jp := range plans {
		jr, err := jp.Execute(ctx)
		if err != nil {
			return nil, err
		}

		runs = append(runs, jr)
	}

	return &WorkflowRun{
		Opts:     wep.opts,
		RunID:    wep.runID,
		Workflow: wep.workflow,
		JobRuns:  runs,
	}, nil
}

// container initializes a new runner container for given workflow execution plan.
func (wep *WorkflowExecutionPlan) container() (*Container, error) {
	var (
		repo     = wep.repo
		workflow = wep.workflow
		opts     = wep.opts
	)

	// initialize base container
	ctr := opts.Container

	// configure internal components
	ctr = ctr.With(dag.Ghx().Source().Binary)
	ctr = ctr.With(dag.ActionsArtifactService().BindAsService)
	ctr = ctr.With(dag.ActionsArtifactcacheService().BindAsService)

	// Configure additional services -- TODO: make it possible use docker socket or remote docker host instead of
	//                                   docker-in-dagger. It costs a lot of time starting docker-in-dagger every time.
	ctr = ctr.With(dag.Docker().WithCacheVolume("gale-docker-cache").BindAsService)

	// GHX specific directory configuration -- TODO: refactor this later to be more generic for runners
	var (
		metadata  = "/home/runner/_temp/gale/metadata"
		actions   = "/home/runner/_temp/gale/actions"
		cacheOpts = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	ctr = ctr.WithEnvVariable("GHX_METADATA_DIR", metadata)
	ctr = ctr.WithMountedCache(metadata, dag.CacheVolume("gale-metadata"), cacheOpts)

	ctr = ctr.WithEnvVariable("GHX_ACTIONS_DIR", actions)
	ctr = ctr.WithMountedCache(actions, dag.CacheVolume("gale-actions"), cacheOpts)

	// Configure repository
	workdir := fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name)

	ctr = ctr.WithMountedDirectory(workdir, repo.Source).WithWorkdir(workdir)
	ctr = ctr.WithEnvVariable("GH_REPO", repo.NameWithOwner)
	ctr = ctr.WithEnvVariable("GITHUB_WORKSPACE", workdir)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY", repo.NameWithOwner)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY_OWNER", repo.Owner)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY_URL", repo.URL)
	ctr = ctr.WithEnvVariable("GITHUB_REF", repo.Ref)
	ctr = ctr.WithEnvVariable("GITHUB_REF_NAME", repo.RefName)
	ctr = ctr.WithEnvVariable("GITHUB_REF_TYPE", repo.RefType)
	ctr = ctr.WithEnvVariable("GITHUB_SHA", repo.SHA)

	// Configure workflow context
	home := filepath.Join("/home/runner/_temp/_gale/runs", wep.runID)

	ctr = ctr.WithMountedDirectory(home, dag.Directory())
	ctr = ctr.WithEnvVariable("GHX_HOME", home)

	// Configure GitHub token if provided
	if opts.Token != nil {
		ctr = ctr.WithSecretVariable("GITHUB_TOKEN", opts.Token)
	}

	// Configure runner debug mode if enabled
	if opts.RunnerDebug {
		ctr = ctr.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	// Configure event
	eventPath := filepath.Join(home, "run", "event.json")

	ctr = ctr.WithEnvVariable("GITHUB_EVENT_NAME", opts.Event)
	ctr = ctr.WithEnvVariable("GITHUB_EVENT_PATH", eventPath)
	ctr = ctr.WithMountedFile(eventPath, opts.EventFile)

	// Configure workflow
	ctr = ctr.WithEnvVariable("GITHUB_RUN_ID", wep.runID)
	ctr = ctr.WithEnvVariable("GITHUB_RUN_NUMBER", "1")
	ctr = ctr.WithEnvVariable("GITHUB_RUN_ATTEMPT", "1")
	ctr = ctr.WithEnvVariable("GITHUB_RETENTION_DAYS", "90")

	path := filepath.Join(home, "run", "workflow.yaml")

	ctr = ctr.WithMountedFile(path, workflow.Src)
	ctr = ctr.WithEnvVariable("GHX_WORKFLOW", workflow.Name)
	ctr = ctr.WithEnvVariable("GITHUB_WORKFLOW", workflow.Name)
	ctr = ctr.WithEnvVariable("GITHUB_WORKFLOW_REF", workflow.Ref)
	ctr = ctr.WithEnvVariable("GITHUB_WORKFLOW_SHA", workflow.SHA)

	// Finalize container configuration
	ctr = ctr.WithEnvVariable("GALE_CONFIGURED", "true")

	return ctr, nil
}

func (wep *WorkflowExecutionPlan) PlanJobs() ([]*JobExecutionPlan, error) {
	var (
		jobs    = make(map[string]*JobExecutionPlan, len(wep.workflow.Jobs))
		plans   = make([]*JobExecutionPlan, 0, len(wep.workflow.Jobs))
		visited = make(map[string]bool)

		visitFn func(name string) error
	)

	// initialized base container
	ctr, err := wep.container()
	if err != nil {
		return nil, err
	}

	// initialize job execution plans
	for _, job := range wep.workflow.Jobs {
		jobs[job.Name] = &JobExecutionPlan{
			runID:  job.JobID,
			job:    job,
			ctr:    ctr,
			parent: wep,
			needs:  make(map[string]*JobExecutionPlan),
		}
	}

	visitFn = func(name string) error {
		if visited[name] {
			return nil
		}

		plan, exist := jobs[name]
		if !exist {
			return fmt.Errorf("job %s not found", name)
		}

		visited[name] = true

		for _, dependency := range plan.job.Needs {
			if err := visitFn(dependency); err != nil {
				return err
			}

			// add dependency to the job plan
			plan.needs[dependency] = jobs[dependency]
		}

		// add job plan to the workflow execution plan
		plans = append(plans, plan)

		return nil
	}

	if wep.opts.Job != "" {
		if err := visitFn(wep.opts.Job); err != nil {
			return nil, err
		}
	} else {
		for _, job := range wep.workflow.Jobs {
			if err := visitFn(job.Name); err != nil {
				return nil, err
			}
		}
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf("failed to find %s job in the workflow", wep.opts.Job)
	}

	return plans, nil
}

type JobExecutionPlan struct {
	// run id of the workflow run.
	runID string

	// the job to run.
	job Job

	// the container for this job run.
	ctr *Container

	// the workflow execution plan for this job run.
	parent *WorkflowExecutionPlan

	// the job execution plans that this job depends on.
	needs map[string]*JobExecutionPlan
}

func (jep *JobExecutionPlan) Execute(ctx context.Context) (*JobRun, error) {
	var (
		home    = filepath.Join("/home/runner/_temp/_gale/runs", jep.parent.runID)
		current = filepath.Join(home, "run/jobs", jep.job.JobID)
		stdout  = filepath.Join(current, "job_run.log")
	)

	// base container
	ctr, err := jep.parent.container()
	if err != nil {
		return nil, err
	}

	// assign container with specific job id to new container to separate each job to its own container
	ctr = ctr.WithEnvVariable("GHX_JOB", jep.runID)

	// mount job dependencies to the container
	for _, need := range jep.needs {
		path := filepath.Join(home, "run/jobs", need.job.JobID)

		ctr = ctr.WithMountedDirectory(path, need.ctr.Directory(path))
	}

	// execute the workflow
	ctr = ctr.WithMountedDirectory(current, dag.Directory()) // just to make sure the directory exists
	ctr = ctr.WithExec(
		[]string{"ghx"},
		ContainerWithExecOpts{
			SkipEntrypoint:                true,
			RedirectStdout:                stdout,
			RedirectStderr:                stdout,
			ExperimentalPrivilegedNesting: true,
			InsecureRootCapabilities:      false,
		},
	)

	ctr, err = ctr.Sync(ctx)
	if err != nil {
		return nil, err
	}

	// update job execution plan with the container to make it accessible for the dependent jobs to mount output files
	jep.ctr = ctr

	return &JobRun{
		Job:    jep.job,
		Ctr:    ctr,
		Data:   ctr.Directory(current),
		Report: ctr.Directory(current).File("job_run.json"),
		Log:    ctr.File(stdout),
	}, nil
}
