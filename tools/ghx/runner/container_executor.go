package runner

import (
	"context"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/log"
	"github.com/aweris/gale/tools/ghx/actions"
	"github.com/aweris/gale/tools/ghx/expression"
)

type ContainerExecutor struct {
	container  *dagger.Container    // container is the container to execute
	stepID     string               // stepID is the ID of the step
	entrypoint string               // entrypoint is the entrypoint of the container
	args       []string             // args is the arguments of the container
	env        map[string]string    // env to pass to the command as environment variables
	ec         *actions.ExprContext // ec is the expression context to evaluate the github expressions
}

func NewContainerExecutorFromStepDocker(sd *StepDocker) *ContainerExecutor {
	env := make(map[string]string)

	// add step level environment variables
	for k, v := range sd.Step.Environment {
		env[k] = v
	}

	return &ContainerExecutor{
		stepID:     sd.Step.ID,
		entrypoint: sd.Step.With["entrypoint"],
		args:       []string{sd.Step.With["args"]},
		env:        env,
		ec:         sd.runner.context,
		container:  sd.container,
	}
}

func (c *ContainerExecutor) Execute(ctx context.Context) error {
	entrypoint := c.entrypoint

	if entrypoint != "" {
		res := expression.NewString(entrypoint)

		// evaluate the expression
		entrypoint, err := res.Eval(c.ec)
		if err != nil {
			log.Errorf("failed to evaluate value", "error", err.Error(), "entrypoint", entrypoint)

			return err
		}

		log.Debugf("entrypoint evaluated", "original", c.entrypoint, "evaluated", entrypoint)

		c.entrypoint = entrypoint
	}

	var args []string

	for _, arg := range c.args {
		str := expression.NewString(arg)

		// evaluate the expression
		res, err := str.Eval(c.ec)
		if err != nil {
			log.Errorf("failed to evaluate value", "error", err.Error(), "arg", arg)

			return err
		}

		log.Debugf("arg evaluated", "original", arg, "evaluated", res)

		args = append(args, strings.Split(res, " ")...) // TODO: this is not correct, we need to parse the string as shell does
	}

	if entrypoint != "" {
		c.container = c.container.WithEntrypoint([]string{entrypoint})
	}

	if len(args) > 0 {
		c.container = c.container.WithExec(args)
	}

	for k, v := range c.env {
		c.container = c.container.WithEnvVariable(k, v)
	}

	_, err := c.container.Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}
