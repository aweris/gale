package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/expression"
)

var _ Executor = new(CmdExecutor)

type CmdExecutor struct {
	args []string          // args to pass to the command
	cp   *CommandProcessor // cp is the command processor to process workflow commands

}

func NewCmdExecutorFromStepAction(sa *StepAction, entrypoint string) *CmdExecutor {
	return &CmdExecutor{
		args: []string{"node", fmt.Sprintf("%s/%s", sa.Action.Path, entrypoint)},
		cp:   NewCommandProcessor(),
	}
}

func NewCmdExecutorFromStepRun(sr *StepRun) *CmdExecutor {
	return &CmdExecutor{
		args: append([]string{sr.Shell}, sr.ShellArgs...),
		cp:   NewCommandProcessor(),
	}
}

func (c *CmdExecutor) Execute(ctx *context.Context) error {
	// variable provider for the container executor. It returns expression.VariableProvider for the container executor
	// based on step being executed. If the step is an action, it returns a variable provider that contains the inputs
	// of the action. Otherwise, it returns the main context as the variable provider.
	vp := ctx.GetVariableProvider()

	//nolint:gosec // this is a command executor, we need to execute the command as it is
	cmd := exec.Command(c.args[0], c.args[1:]...)

	envMap := make(map[string]string)

	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	efs, err := NewLocalEnvironmentFiles(filepath.Join(ctx.Runner.Temp, "env_files"))
	if err != nil {
		return err
	}

	envMap[EnvFileNameGithubEnv] = efs.Env.Path()
	envMap[EnvFileNameGithubPath] = efs.Path.Path()
	envMap[EnvFileNameGithubOutput] = efs.Outputs.Path()
	envMap[EnvFileNameGithubStepSummary] = efs.StepSummary.Path()

	// update the expression context with the environment files
	ctx.WithGithubEnv(efs.Env.Path()).WithGithubPath(efs.Path.Path())
	defer func() {
		ctx.WithoutGithubEnv().WithoutGithubPath()
	}()

	// add environment variables

	if ctx.Execution.CurrentAction != nil {
		for k, v := range ctx.Execution.StepRun.Step.With {
			envMap[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v
		}

		// add default values for inputs that are not defined in the step config
		for k, v := range ctx.Execution.CurrentAction.Meta.Inputs {
			if _, ok := ctx.Execution.StepRun.Step.With[k]; ok {
				continue
			}

			if v.Default == "" {
				continue
			}

			envMap[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v.Default
		}
	}

	// add step state to the environment
	for k, v := range ctx.Steps[ctx.Execution.StepRun.Step.ID].State {
		envMap[fmt.Sprintf("STATE_%s", k)] = v
	}

	// add step level environment variables
	for k, v := range ctx.Env {
		envMap[k] = v
	}

	env := os.Environ()

	for k, v := range envMap {
		// convert value to Evaluable String type
		str := expression.NewString(v)

		// evaluate the expression
		res := str.Eval(vp)

		log.Debugf("Environment variable evaluated", "key", k, "value", v, "evaluated", res)

		env = append(env, fmt.Sprintf("%s=%s", k, res))
	}

	cmd.Env = env

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			output := scanner.Text()

			if err := c.cp.ProcessOutput(ctx, output); err != nil {
				log.Errorf("failed to process output", "output", output, "error", err)
			}
		}
	}()

	waitErr := cmd.Wait()

	if err := efs.Process(ctx); err != nil {
		return err
	}

	return waitErr
}
