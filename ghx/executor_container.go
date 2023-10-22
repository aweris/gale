package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/expression"
)

var _ Executor = new(ContainerExecutor)

type ContainerExecutor struct {
	container  *dagger.Container // container is the container to execute
	entrypoint string            // entrypoint is the entrypoint of the container
	args       []string          // args is the arguments of the container
	cp         *CommandProcessor // cp is the command processor to process workflow commands
}

func NewContainerExecutorFromStepDocker(sd *StepDocker) *ContainerExecutor {
	return &ContainerExecutor{
		entrypoint: sd.Step.With["entrypoint"],
		args:       []string{sd.Step.With["args"]},
		cp:         NewEnvCommandsProcessor(),
		container:  sd.container,
	}
}

func NewContainerExecutorFromStepAction(sa *StepAction, entrypoint string) *ContainerExecutor {
	return &ContainerExecutor{
		entrypoint: entrypoint,
		args:       sa.Action.Meta.Runs.Args,
		cp:         NewEnvCommandsProcessor(),
		container:  sa.container,
	}
}

func (c *ContainerExecutor) Execute(ctx *context.Context) error {
	// variable provider for the container executor. It returns expression.VariableProvider for the container executor
	// based on step being executed. If the step is an action, it returns a variable provider that contains the inputs
	// of the action. Otherwise, it returns the main context as the variable provider.
	vp := ctx.GetVariableProvider()

	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	dir, efs := NewDaggerEnvironmentFiles(filepath.Join(ctx.Runner.Temp, "env_files"), ctx.Dagger.Client)

	c.container = c.container.WithMountedDirectory(filepath.Join(ctx.Runner.Temp, "env_files"), dir).
		WithEnvVariable(EnvFileNameGithubEnv, efs.Env.Path()).
		WithEnvVariable(EnvFileNameGithubPath, efs.Path.Path()).
		WithEnvVariable(EnvFileNameGithubOutput, efs.Outputs.Path()).
		WithEnvVariable(EnvFileNameGithubStepSummary, efs.StepSummary.Path())

	// update the expression context with the environment files
	ctx.WithGithubEnv(efs.Env.Path()).WithGithubPath(efs.Path.Path())
	defer func() {
		ctx.WithoutGithubEnv().WithoutGithubPath()
	}()

	entrypoint := c.entrypoint

	if entrypoint != "" {
		res := expression.NewString(entrypoint)

		// evaluate the expression
		entrypoint := res.Eval(ctx)

		log.Debugf("entrypoint evaluated", "original", c.entrypoint, "evaluated", entrypoint)

		c.entrypoint = entrypoint
	}

	var args []string

	for _, arg := range c.args {
		str := expression.NewString(arg)

		// evaluate the expression
		res := str.Eval(vp)

		log.Debugf("arg evaluated", "original", arg, "evaluated", res)

		args = append(args, strings.Split(res, " ")...) // TODO: this is not correct, we need to parse the string as shell does
	}

	if entrypoint != "" {
		c.container = c.container.WithEntrypoint([]string{entrypoint})
	}

	if len(args) > 0 {
		c.container = c.container.WithExec(args, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	}

	env := make(map[string]string)

	if ctx.Execution.CurrentAction != nil {
		for k, v := range ctx.Execution.CurrentAction.Meta.Runs.Env {
			env[k] = v
		}

		for k, v := range ctx.Execution.StepRun.Step.With {
			env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v
		}

		// add default values for inputs that are not defined in the step config
		for k, v := range ctx.Execution.CurrentAction.Meta.Inputs {
			if _, ok := ctx.Execution.StepRun.Step.With[k]; ok {
				continue
			}

			if v.Default == "" {
				continue
			}

			env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v.Default
		}
	}

	// add step state to the environment
	for k, v := range ctx.Steps[ctx.Execution.StepRun.Step.ID].State {
		env[fmt.Sprintf("STATE_%s", k)] = v
	}

	// add step level environment variables
	for k, v := range ctx.Env {
		env[k] = v
	}

	for k, v := range env {
		c.container = c.container.WithEnvVariable(k, v)
	}

	// TODO: if no args are provided, we need to execute the container with the default entrypoint and args
	//  however this is causing an error since Stdout is looking for last execs output. We need to find a way to
	//  execute the container without execs and get the output.
	out, err := c.container.Stdout(ctx.Context)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		output := scanner.Text()

		if err := c.cp.ProcessOutput(ctx, output); err != nil {
			log.Errorf("failed to process output", "output", output, "error", err)
		}
	}

	return efs.Process(ctx)
}
