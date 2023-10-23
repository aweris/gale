package main

import (
	"fmt"
	"path/filepath"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/expression"
	"github.com/aweris/gale/ghx/task"
)

// Step is
var _ Step = new(StepRun)

// StepRun is a step that runs a job.
type StepRun struct {
	Step      model.Step
	Shell     string   // Shell is the shell to use to run the script.
	ShellArgs []string // ShellArgs are the arguments to pass to the shell.
	Path      string   // Path is the script path to run.

}

func (s *StepRun) condition() task.ConditionalFn {
	return func(ctx *context.Context) (bool, model.Conclusion, error) {
		return evalCondition(s.Step.If, ctx)
	}
}

const (
	extSH   = ".sh"
	extPY   = ".py"
	extPWSH = ".ps1"
)

func (s *StepRun) main() task.RunFn {
	return func(ctx *context.Context) (model.Conclusion, error) {
		var (
			pre   string
			pos   string
			args  []string
			shell = s.Step.Shell
		)

		dir, err := ctx.GetStepRunPath()
		if err != nil {
			return model.ConclusionFailure, err
		}

		path := filepath.Join(dir, "run")

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
			return model.ConclusionFailure, fmt.Errorf("not supported shell: %s", shell)
		}

		// evaluate run script against the expressions
		run := expression.NewString(s.Step.Run).Eval(ctx)

		content := []byte(fmt.Sprintf("%s\n%s\n%s", pre, run, pos))

		err = fs.WriteFile(path, content, 0755)
		if err != nil {
			return model.ConclusionFailure, err
		}

		s.Shell = shell
		s.ShellArgs = args
		s.Path = path

		executor := NewCmdExecutorFromStepRun(s)

		// execute the step
		if err := executor.Execute(ctx); err != nil {
			if s.Step.ContinueOnError {
				// execution failed and the step is configured to continue on error. So, fail the outcome but succeed the
				// conclusion.
				ctx.SetStepResults(model.ConclusionSuccess, model.ConclusionFailure)

				return model.ConclusionSuccess, nil
			}

			// execution failed and the step is not configured to continue on error. So, we need to fail the step.
			ctx.SetStepResults(model.ConclusionFailure, model.ConclusionFailure)

			return model.ConclusionFailure, err
		}

		// update the step outputs
		ctx.SetStepResults(model.ConclusionSuccess, model.ConclusionSuccess)

		return model.ConclusionSuccess, nil
	}
}
