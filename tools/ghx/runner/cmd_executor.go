package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aweris/gale/tools/ghx/actions"
)

type CmdExecutor struct {
	args []string             // args to pass to the command
	env  map[string]string    // env to pass to the command as environment variables
	ec   *actions.ExprContext // ec is the expression context to evaluate the github expressions
}

func NewCmdExecutorFromStepAction(sa *StepAction, entrypoint string) *CmdExecutor {
	env := make(map[string]string)

	for k, v := range sa.Step.With {
		env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v
	}

	// add default values for inputs that are not defined in the step config
	for k, v := range sa.Action.Meta.Inputs {
		if _, ok := sa.Step.With[k]; ok {
			continue
		}

		if v.Default == "" {
			continue
		}

		env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v.Default
	}

	return &CmdExecutor{
		args: []string{"node", fmt.Sprintf("%s/%s", sa.Action.Path, entrypoint)},
		env:  env,
		ec:   &actions.ExprContext{}, // empty expression context TODO: provide the actual context
	}
}

func NewCmdExecutorFromStepRun(sr *StepRun) *CmdExecutor {
	args := []string{sr.Shell}

	args = append(args, sr.ShellArgs...)
	args = append(args, sr.Path)

	return &CmdExecutor{
		args: args,
		env:  make(map[string]string),
		ec:   &actions.ExprContext{}, // empty expression context TODO: provide the actual context
	}
}

func (c *CmdExecutor) Execute(_ context.Context) error {
	cmd := exec.Command(c.args[0], c.args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// add environment variables

	env := os.Environ()

	for k, v := range c.env {
		// convert value to Evaluable String type
		str := actions.NewString(v)

		// evaluate the expression
		res, err := str.Eval(c.ec)
		if err != nil {
			return fmt.Errorf("failed to evaluate default value for input %s: %v", k, err)
		}

		env = append(env, fmt.Sprintf("%s=%s", k, res))
	}

	cmd.Env = env

	return cmd.Run()
}
