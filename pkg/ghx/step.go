package ghx

import (
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/log"
)

// Step is an internal interface that defines contract for steps.
type Step interface {
	// condition returns the function that checks if the main execution condition is met.
	condition() TaskConditionalFn

	// main returns the function that executes the main execution logic.
	main() TaskRunFn
}

// SetupHook is the interface that defines contract for steps capable of performing a setup task.
type SetupHook interface {
	// setup returns the function that sets up the step before execution.
	setup() TaskRunFn
}

// PreHook is the interface that defines contract for steps capable of performing a pre execution task.
type PreHook interface {
	// preCondition returns the function that checks if the pre execution condition is met.
	preCondition() TaskConditionalFn

	// pre returns the function that executes the pre execution logic just before the main execution.
	pre() TaskRunFn
}

// PostHook is the interface that defines contract for steps capable of performing a post execution task.
type PostHook interface {
	// postCondition returns the function that checks if the post execution condition is met.
	postCondition() TaskConditionalFn

	// post returns the function that executes the post execution logic just after the main execution.
	post() TaskRunFn
}

// NewStep creates a new step from the given step configuration.
func NewStep(s core.Step) (Step, error) {
	var step Step

	switch s.Type() {
	case core.StepTypeAction:
		step = &StepAction{Step: s}
	case core.StepTypeRun:
		step = &StepRun{Step: s}
	case core.StepTypeDocker:
		step = &StepDocker{Step: s}
	default:
		return nil, fmt.Errorf("unknown step type: %s", s.Type())
	}

	return step, nil
}

var (
	_ Step      = new(StepAction)
	_ PreHook   = new(StepAction)
	_ PostHook  = new(StepAction)
	_ SetupHook = new(StepAction)
)

// FIXME: Currently Steps and context holds the shared information. This is not a good design. We should refactor this
//  to make sure that steps and context are not sharing the same information.

// StepAction is a step that runs an action.
type StepAction struct {
	container *dagger.Container
	Step      core.Step
	Action    core.CustomAction
}

func (s *StepAction) setup() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		ca, err := core.LoadActionFromSource(ctx.Context, s.Step.Uses)
		if err != nil {
			return core.ConclusionFailure, err
		}

		// update the step action with the loaded action
		s.Action = *ca

		log.Info(fmt.Sprintf("Download action repository '%s'", s.Step.Uses))

		if s.Action.Meta.Runs.Using == core.ActionRunsUsingDocker {
			var (
				image        = ca.Meta.Runs.Image
				workspace    = ctx.Github.Workspace
				workspaceDir = config.Client().Host().Directory(workspace)
			)

			switch {
			case image == "Dockerfile":
				s.container = config.Client().Container().Build(ca.Dir)
			case strings.HasPrefix(image, "docker://"):
				s.container = config.Client().Container().From(strings.TrimPrefix(image, "docker://"))
			default:
				// This should never happen. Adding it for safety.
				return core.ConclusionFailure, fmt.Errorf("invalid docker image: %s", image)
			}

			// add repository to the container
			s.container = s.container.WithMountedDirectory(workspace, workspaceDir).WithWorkdir(workspace)
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) preCondition() TaskConditionalFn {
	return func(ctx *gctx.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PreCondition()
		if !run {
			return false, "", nil
		}

		return evalStepCondition(condition, ctx)
	}
}

func (s *StepAction) pre() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.PreEntrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			executor = NewCmdExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.Pre)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		if err := executor.Execute(ctx); err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) condition() TaskConditionalFn {
	return func(ctx *gctx.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, ctx)
	}
}

func (s *StepAction) main() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.Entrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			executor = NewCmdExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.Main)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		if err := executor.Execute(ctx); err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) postCondition() TaskConditionalFn {
	return func(ctx *gctx.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PostCondition()
		if !run {
			return false, "", nil
		}

		return evalStepCondition(condition, ctx)
	}
}

func (s *StepAction) post() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.PostEntrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			executor = NewCmdExecutorFromStepAction(ctx, s, s.Action.Meta.Runs.Post)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		if err := executor.Execute(ctx); err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

var _ Step = new(StepRun)

// StepRun is a step that runs a job.
type StepRun struct {
	Step      core.Step
	Shell     string   // Shell is the shell to use to run the script.
	ShellArgs []string // ShellArgs are the arguments to pass to the shell.
	Path      string   // Path is the script path to run.

}

func (s *StepRun) condition() TaskConditionalFn {
	return func(ctx *gctx.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, ctx)
	}
}

const (
	extSH   = ".sh"
	extPY   = ".py"
	extPWSH = ".ps1"
)

func (s *StepRun) main() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		var (
			pre   string
			pos   string
			args  []string
			shell = s.Step.Shell
		)

		jrPath, err := ctx.GetJobRunPath()
		if err != nil {
			return core.ConclusionFailure, err
		}

		path := filepath.Join(jrPath, "scripts", s.Step.ID, "run.sh")

		// TODO: add support and test for "pwsh" shell. This is not supported for now because we don't have a pwsh image,
		//  and we don't have a way to test it for now. We'll add support for it later after making sure that it works.

		// set the shell and shell args according to the shell type. Windows is not supported platform for now. So, we
		// don't need to handle shell type for windows.
		//
		// Docs: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsshell
		// Ref: https://github.com/actions/runner/blob/efffbaeabc6d53c4c1ec05b11cea58331ff38e3c/src/Runner.Worker/Handlers/ScriptHandlerHelpers.cs
		// Ref: https://github.com/actions/runner/blob/efffbaeabc6d53c4c1ec05b11cea58331ff38e3c/src/Runner.Worker/Handlers/ScriptHandler.cs
		switch shell {
		case "":
			path += extSH
			shell = "bash"
			args = []string{"-e", path}
		case "bash":
			path += extSH
			args = []string{"--noprofile", "--norc", "-e", "-o", "pipefail", path}
		case "python":
			path += extPY
			args = []string{path}
		case "pwsh":
			path += extPWSH
			pre = "$ErrorActionPreference = 'stop'"
			pos = "if ((Test-Path -LiteralPath variable:/LASTEXITCODE)) { exit $LASTEXITCODE }"
			args = []string{"-command", fmt.Sprintf(". '%s'", path)}
		case "sh":
			path += extSH
			args = []string{"-e", path}
		default:
			return core.ConclusionFailure, fmt.Errorf("not supported shell: %s", shell)
		}

		// evaluate run script against the expressions
		run, err := expression.NewString(s.Step.Run).Eval(ctx)
		if err != nil {
			return core.ConclusionFailure, err
		}

		content := []byte(fmt.Sprintf("%s\n%s\n%s", pre, run, pos))

		err = fs.WriteFile(path, content, 0755)
		if err != nil {
			return core.ConclusionFailure, err
		}

		s.Shell = shell
		s.ShellArgs = args
		s.Path = path

		cmd := NewCmdExecutorFromStepRun(ctx, s)

		err = cmd.Execute(ctx)
		if err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, err
	}
}

var (
	_ Step      = new(StepDocker)
	_ SetupHook = new(StepDocker)
)

type StepDocker struct {
	container *dagger.Container
	Step      core.Step
}

func (s *StepDocker) setup() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		var (
			image        = strings.TrimPrefix(s.Step.Uses, "docker://")
			workspace    = ctx.Github.Workspace
			workspaceDir = config.Client().Host().Directory(workspace)
		)

		// configure the step container
		s.container = config.Client().
			Container().
			From(image).
			WithMountedDirectory(workspace, workspaceDir).
			WithWorkdir(workspace)

		// TODO: This will be print same log line if the image used multiple times. However, this scenario is not really common and no benefit to fix this scenario for now.
		log.Info(fmt.Sprintf("Pull '%s'", image))

		return core.ConclusionSuccess, nil
	}
}

func (s *StepDocker) condition() TaskConditionalFn {
	return func(ctx *gctx.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, ctx)
	}
}

func (s *StepDocker) main() TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		executor := NewContainerExecutorFromStepDocker(ctx, s)

		err := executor.Execute(ctx)
		if err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

func newTaskPreRunFnForStep(step core.Step) TaskPreRunFn {
	return func(ctx *gctx.Context) error {
		ctx.SetStep(step)

		return nil
	}
}

func newTaskPostRunFnForStep(_ core.Step) TaskPostRunFn {
	return func(ctx *gctx.Context) (err error) {
		ctx.UnsetStep()

		return nil
	}
}
