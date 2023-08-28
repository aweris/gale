package preflight

import (
	"fmt"
	"os"
	"strings"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
)

var _ Task = new(WorkflowLoader)

type WorkflowLoader struct{}

func (c *WorkflowLoader) Name() string {
	return NameWorkflowLoader
}

func (c *WorkflowLoader) Type() TaskType {
	return TaskTypeLoad
}

func (c *WorkflowLoader) DependsOn() []string {
	return []string{NameDaggerCheck, NameRepoLoader}
}

func (c *WorkflowLoader) Run(ctx *Context, opts Options) Result {
	var msg []Message

	repo := ctx.Repo

	workflows, err := repo.LoadWorkflows(ctx.Context, core.RepositoryLoadWorkflowOpts{WorkflowsDir: opts.WorkflowsDir})
	if err != nil {
		return Result{
			Status: Failed,
			Messages: []Message{
				{Level: Error, Content: fmt.Sprintf("Load workflows failed: %s", err.Error())},
			},
		}
	}

	ctx.Workflows = workflows

	if opts.WorkflowsDir == "" {
		msg = append(msg, Message{Level: Info, Content: "Workflows are loaded from .github/workflows"})
	} else {
		msg = append(msg, Message{Level: Info, Content: fmt.Sprintf("Workflows are loaded from %s", opts.WorkflowsDir)})
	}

	workflow, ok := workflows[opts.Workflow]
	if !ok {
		msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Workflow %s is not found", opts.Workflow)})

		return Result{Status: Failed, Messages: msg}
	}

	ctx.Workflow = workflow

	msg = append(msg, Message{Level: Info, Content: fmt.Sprintf("Workflow %s is loaded", opts.Workflow)})

	job, ok := workflow.Jobs[opts.Job]
	if !ok {
		msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Job %s is not found", opts.Job)})

		return Result{Status: Failed, Messages: msg}
	}

	ctx.Job = &job

	msg = append(msg, Message{Level: Info, Content: fmt.Sprintf("Job %s is loaded", opts.Job)})

	// TODO: find a better way to do this. Maybe pass target directory as an option or add clean up function to the context and call it at the end.
	dir, err := os.MkdirTemp("", "ghx-home")
	if err != nil {
		msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Failed to create temporary GHX home directory: %s", err.Error())})

		return Result{Status: Failed, Messages: msg}
	}

	config.SetGhxHome(dir)

	for _, step := range job.Steps {
		// check if the step is an action
		switch step.Type() {
		case core.StepTypeAction:
			if ctx.CustomActions == nil {
				ctx.CustomActions = make(map[string]*core.CustomAction)
			}

			// check if the action is already loaded
			if ctx.CustomActions[step.Uses] != nil {
				continue
			}

			// load the action
			ca, err := core.LoadActionFromSource(ctx.Context, step.Uses)
			if err != nil {
				msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Load action %s failed: %s", step.Uses, err.Error())})

				return Result{Status: Failed, Messages: msg}
			}

			// add the action to the list
			ctx.CustomActions[step.Uses] = ca

			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Action %s is used in the workflow", step.Uses)})

			// check if the action uses a Docker image
			if ca.Meta.Runs.Image != "" && strings.HasPrefix(ca.Meta.Runs.Image, "docker://") {
				// trim docker:// prefix
				image := strings.TrimPrefix(ca.Meta.Runs.Image, "docker://")

				// check if map is initialized
				if ctx.DockerImages == nil {
					ctx.DockerImages = make(map[string]bool)
				}

				// check if the image is already used
				if ctx.DockerImages[image] {
					continue
				}

				// add the image to the list
				ctx.DockerImages[image] = true

				msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Docker image %s is used in the workflow", image)})
			}
		case core.StepTypeDocker:
			// trim docker:// prefix
			image := strings.TrimPrefix(step.Uses, "docker://")

			// check if map is initialized
			if ctx.DockerImages == nil {
				ctx.DockerImages = make(map[string]bool)
			}

			// check if the image is already used
			if ctx.DockerImages[image] {
				continue
			}

			// add the image to the list
			ctx.DockerImages[image] = true

			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Docker image %s is used in the workflow", image)})
		case core.StepTypeRun:
			// check if map is initialized
			if ctx.Shells == nil {
				ctx.Shells = make(map[string]bool)
			}

			// empty shell means the default shell. skip it.
			if step.Shell == "" {
				ctx.Shells["bash"] = true // default shell is bash
				continue
			}

			// check if the shell is already used
			if ctx.Shells[step.Shell] {
				continue
			}

			// add the shell to the list

			ctx.Shells[step.Shell] = true

			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Shell %s is used in the workflow", step.Shell)})
		default:
			msg = append(msg, Message{Level: Warning, Content: fmt.Sprintf("Step %s is not supported", step.Type())})
		}
	}

	return Result{Status: Passed, Messages: msg}
}
