package steps

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/google/uuid"

	"github.com/aweris/gale/model"
	"github.com/aweris/gale/runner/container"
	"github.com/aweris/gale/runner/state"
)

type Handler struct {
	ch    *container.Handler
	state *state.StepRunState
}

func NewHandler(ch *container.Handler, state *state.StepRunState) *Handler {
	return &Handler{ch: ch, state: state}
}

func (h *Handler) WithAction(ctx context.Context, source string) error {
	// TODO: double check if this is the right way to do this
	action, err := model.LoadActionFromSource(ctx, h.state.Client, source)
	if err != nil {
		return err
	}

	h.state.Action = action
	h.state.ActionPath = fmt.Sprintf("/home/runner/_temp/actions/%s", uuid.New())

	h.ch.WithDirectory(ctx, h.state.ActionPath, action.Directory, dagger.ContainerWithDirectoryOpts{Owner: "runner:runner"})

	return nil
}

type ExecStepStatus string

const (
	StatusSucceeded ExecStepStatus = "succeeded"
	StatusSkipped   ExecStepStatus = "skipped"
	StatusFailed    ExecStepStatus = "failed"
)

func (h *Handler) ExecStep(ctx context.Context, stage model.ActionStage) error {
	// Apply step state as environment variable to the container
	h.ch.WithEnv(h.state.GetStateEnv())

	// Apply step environment variable to the container
	h.ch.WithEnv(h.state.Step.Environment)

	// TODO: add check for the step type for shell, docker, etc. and publish the appropriate event.
	// For now, we only support actions and run steps
	switch h.state.Step.Type() {
	case model.StepTypeAction:
		h.ExecStepAction(ctx, stage)
	case model.StepTypeRun:
		h.ExecStepRun(ctx, stage)
	default:
		return fmt.Errorf("unsupported step type")
	}

	// Remove step environment variable to the container. If there is fallback environment variable, apply them
	// to the container instead of removing them. This is to make sure that the container has the right environment
	h.ch.WithoutEnv(h.state.Step.Environment, h.state.GetStepFallbackEnvs()...)

	// Remove step state from environment variable to the container
	h.ch.WithoutEnv(h.state.GetStateEnv())

	return nil
}

func (h *Handler) ExecStepAction(ctx context.Context, stage model.ActionStage) (ExecStepStatus, error) {
	var runs string

	switch stage {
	case model.ActionStagePre:
		runs = h.state.Action.Runs.Pre
	case model.ActionStageMain:
		runs = h.state.Action.Runs.Main
	case model.ActionStagePost:
		runs = h.state.Action.Runs.Post
	default:
		return StatusFailed, fmt.Errorf("unknown stage %s", stage)
	}

	// if runs is empty for pre or post, this is a no-op step
	if runs == "" && stage != model.ActionStageMain {
		return StatusSkipped, nil
	}

	if runs == "" && stage == model.ActionStageMain {
		h.state.Result.Conclusion = model.StepStatusFailure
		h.state.Result.Outcome = model.StepStatusFailure

		return StatusFailed, fmt.Errorf("no runs for step %s", h.state.Step.ID)
	}

	// TODO: check if conditions

	// TODO: check if we can add this to as pipeline name to the container
	fmt.Printf("%s Run %s", stage, h.state.Step.Uses)

	// Apply input environment variable to the container
	h.ch.WithEnv(h.state.GetInputEnv())

	args := []string{"node", fmt.Sprintf("%s/%s", h.state.ActionPath, runs)}

	result, err := h.ch.WithExec(ctx, args, container.WithExecOpts{Execute: true, Strace: true})

	if result.Stdout != "" {
		h.state.ExportLogArtifact(fmt.Sprintf("%s-stdout.log", stage), result.Stdout)
	}

	if result.Strace != "" {
		h.state.ExportLogArtifact(fmt.Sprintf("%s-strace.log", stage), result.Strace)
	}

	if err != nil {
		return StatusFailed, err
	}

	// process github workflow commands
	h.ProcessGithubWorkflowCommands(ctx, result.Stdout)

	// Remove input environment variable to the container
	h.ch.WithoutEnv(h.state.GetInputEnv())

	return StatusSucceeded, nil
}

func (h *Handler) ExecStepRun(ctx context.Context, stage model.ActionStage) (ExecStepStatus, error) {
	// step run only happens for main stage
	if stage != model.ActionStageMain {
		return StatusSkipped, nil
	}

	path := fmt.Sprintf("/home/runner/_temp/scripts/%s", uuid.New())

	h.ch.WithNewFile(
		ctx,
		path,
		dagger.ContainerWithNewFileOpts{
			Contents:    fmt.Sprintf("#!/bin/bash\n%s", h.state.Step.Run),
			Permissions: 0755,
			Owner:       "runner:runner",
		},
	)
	exec := []string{"bash", "--noprofile", "--norc", "-e", "-o", "pipefail", path}

	result, err := h.ch.WithExec(ctx, exec, container.WithExecOpts{Execute: true, Strace: true})

	if result.Stdout != "" {
		h.state.ExportLogArtifact(fmt.Sprintf("%s-stdout.log", stage), result.Stdout)
	}

	if result.Strace != "" {
		h.state.ExportLogArtifact(fmt.Sprintf("%s-strace.log", stage), result.Strace)
	}

	if err != nil {
		return StatusFailed, err
	}

	// process github workflow commands
	h.ProcessGithubWorkflowCommands(ctx, result.Stdout)

	return StatusSucceeded, nil
}

// TODO maybe we need to move part of this as script inside of the container to make it nicer with dagger logs

func (h *Handler) ProcessGithubWorkflowCommands(ctx context.Context, stdout string) error {
	scanner := bufio.NewScanner(strings.NewReader(stdout))

	// Loop through each line and process it
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		isCommand, command := model.ParseCommand(line)
		if !isCommand {
			continue
		}

		// Only handle commands that are make modifications to the environment. Logging commands are handled by the logger.
		switch command.Name {
		case "set-env":
			h.ch.WithEnvVariable(command.Parameters["name"], command.Value)
		case "set-output":
			h.state.SetOutput(command.Parameters["name"], command.Value)
		case "save-state":
			h.state.SaveState(command.Parameters["name"], command.Value)
		case "add-mask":
			fmt.Printf("add-mask: %s\n", command.Value)
		case "add-matcher":
			fmt.Printf("add-matcher: %s\n", command.Value)
		case "add-path": //TODO: make it nicer
			if err := h.ch.WithPath(ctx, command.Value); err != nil {
				return err
			}
		}
	}

	return nil
}
