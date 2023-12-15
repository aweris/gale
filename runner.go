package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Runner struct {
	RunnerOpts *RunnerOpts
	EventOpts  *EventOpts
	SecretOpts *SecretOpts
	Repo       *RepoInfo
	Workflow   *Workflow
}

func NewRunner(
	repo *RepoInfo,
	workflow *Workflow,
	runnerOpts *RunnerOpts,
	eventOpts *EventOpts,
	secretOpts *SecretOpts,
) *Runner {
	return &Runner{
		RunnerOpts: runnerOpts,
		EventOpts:  eventOpts,
		SecretOpts: secretOpts,
		Repo:       repo,
		Workflow:   workflow,
	}
}

type RunnerContainer struct {
	RunID string
	Ctr   *Container
}

func (r *Runner) Container(runID string) (*RunnerContainer, error) {
	var (
		repo     = r.Repo
		workflow = r.Workflow
		ctr      = r.RunnerOpts.Ctr
	)

	// configure internal components
	ctr = ctr.With(dag.Ghx().Binary)
	ctr = ctr.With(dag.ActionsArtifactService().BindAsService)
	ctr = ctr.With(dag.ActionsArtifactcacheService().BindAsService)

	// configure docker using docker host when docker is enabled and dind is not enabled
	if r.RunnerOpts.UseNativeDocker && !r.RunnerOpts.UseDind {
		switch {
		case strings.HasPrefix(r.RunnerOpts.DockerHost, "unix://"):
			dh := strings.TrimPrefix(r.RunnerOpts.DockerHost, "unix://")

			ctr = ctr.WithUnixSocket("/var/run/docker.sock", dag.Host().UnixSocket(dh))
			ctr = ctr.WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock")
		case strings.HasPrefix(r.RunnerOpts.DockerHost, "tcp://"):
			ctr = ctr.WithEnvVariable("DOCKER_HOST", r.RunnerOpts.DockerHost)
		default:
			return nil, fmt.Errorf("unsupported docker host: %s", r.RunnerOpts.DockerHost)
		}
	}

	// configure docker-in-dagger service
	if r.RunnerOpts.UseDind {
		ctr = ctr.With(dag.Docker().BindAsService)
	}

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
	home := filepath.Join("/home/runner/_temp/_gale/runs", runID)

	ctr = ctr.WithMountedDirectory(home, dag.Directory())
	ctr = ctr.WithEnvVariable("GHX_HOME", home)

	// Configure GitHub token if provided
	if r.SecretOpts.Token != nil {
		ctr = ctr.WithSecretVariable("GITHUB_TOKEN", r.SecretOpts.Token)
	}

	// Configure runner debug mode if enabled
	if r.RunnerOpts.Debug {
		ctr = ctr.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	// Configure event
	eventPath := filepath.Join(home, "run", "event.json")

	ctr = ctr.WithEnvVariable("GITHUB_EVENT_NAME", r.EventOpts.Name)
	ctr = ctr.WithEnvVariable("GITHUB_EVENT_PATH", eventPath)
	ctr = ctr.WithMountedFile(eventPath, r.EventOpts.File)

	// Configure workflow
	ctr = ctr.WithEnvVariable("GITHUB_RUN_ID", runID)
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

	return &RunnerContainer{RunID: runID, Ctr: ctr}, nil
}

func (rc *RunnerContainer) RunJob(ctx context.Context, job *Job, conclusion string, needs ...*JobRun) (jr *JobRun,
	err error) {
	var (
		home    = filepath.Join("/home/runner/_temp/_gale/runs", rc.RunID)
		current = filepath.Join(home, "run/jobs", job.JobID)
		stdout  = filepath.Join(current, "job_run.log")
	)

	// base container
	ctr := rc.Ctr

	// assign container with specific job id to new container to separate each job to its own container
	ctr = ctr.WithEnvVariable("GHX_JOB", job.JobID)

	// configure container with actual workflow conclusion status
	ctr = ctr.WithEnvVariable("GHX_WORKFLOW_CONCLUSION", conclusion)

	// mount data directories of the jobs this job depends on
	for _, need := range needs {
		path := filepath.Join(home, "run/jobs", need.Job.JobID)

		ctr = ctr.WithMountedDirectory(path, need.Data)
	}

	// ensure the job directory exists
	ctr = ctr.WithMountedDirectory(current, dag.Directory())

	// enable focus mode for the job
	ctr = ctr.WithFocus()

	// execute the job
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

	// disable focus mode after the job is executed
	ctr = ctr.WithoutFocus()

	ctr, err = ctr.Sync(ctx)
	if err != nil {
		return nil, err
	}

	report, err := parseJobRunReport(ctx, ctr.Directory(current).File("job_run.json"))
	if err != nil {
		return nil, err
	}

	jr = &JobRun{
		Job:     job,
		Ctr:     ctr,
		Data:    ctr.Directory(current),
		Report:  report,
		LogFile: ctr.Directory(current).File("job_run.log"),
	}

	return jr, nil
}
