package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/log"
	"github.com/aweris/gale/tools/ghx/actions"
	"github.com/aweris/gale/tools/ghx/expression"
)

var _ Executor = new(ContainerExecutor)

type ContainerExecutor struct {
	container  *dagger.Container       // container is the container to execute
	stepID     string                  // stepID is the ID of the step
	entrypoint string                  // entrypoint is the entrypoint of the container
	args       []string                // args is the arguments of the container
	env        map[string]string       // env to pass to the command as environment variables
	ec         *actions.ExprContext    // ec is the expression context to evaluate the github expressions
	dec        *actions.ExprContext    // dec is the dagger expression context to evaluate action meta files. TODO: temporary solution, we need to find a better way to do this
	commands   []*core.WorkflowCommand // commands is the list of commands that are executed in the step
	envFiles   *EnvironmentFiles       // envFiles contains temporary files that can be used to perform certain actions
}

func NewContainerExecutorFromStepDocker(sd *StepDocker) *ContainerExecutor {
	return &ContainerExecutor{
		stepID:     sd.Step.ID,
		entrypoint: sd.Step.With["entrypoint"],
		args:       []string{sd.Step.With["args"]},
		env:        sd.Step.Environment,
		dec:        sd.runner.context,
		ec:         sd.runner.context,
		container:  sd.container,
	}
}

func NewContainerExecutorFromStepAction(sa *StepAction, entrypoint string) *ContainerExecutor {
	// get context as value to avoid changing the original context. We need to do this because we are going to change
	// inputs context and we don't want to change the original context.
	dec := *sa.runner.context

	env := make(map[string]string)

	for k, v := range sa.Action.Meta.Runs.Env {
		env[k] = v
	}

	for k, v := range sa.Step.With {
		env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v
		dec.Inputs[k] = v
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
		dec.Inputs[k] = v.Default
	}

	// add step state to the environment
	for k, v := range sa.runner.context.Steps[sa.Step.ID].State {
		env[fmt.Sprintf("STATE_%s", k)] = v
	}

	// add step level environment variables
	for k, v := range sa.Step.Environment {
		env[k] = v
	}

	var args []string

	args = append(args, sa.Action.Meta.Runs.Args...)

	return &ContainerExecutor{
		stepID:     sa.Step.ID,
		entrypoint: entrypoint,
		args:       args,
		env:        env,
		ec:         sa.runner.context,
		dec:        &dec,
		container:  sa.container,
	}
}

func (c *ContainerExecutor) Execute(ctx context.Context) error {
	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	err := c.loadEnvFiles()
	if err != nil {
		return err
	}
	defer c.unloadEnvFiles()

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
		res, err := str.Eval(c.dec)
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

	// TODO: if no args are provided, we need to execute the container with the default entrypoint and args
	//  however this is causing an error since Stdout is looking for last execs output. We need to find a way to
	//  execute the container without execs and get the output.
	out, err := c.container.Stdout(ctx)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		output := scanner.Text()

		isCommand, command := core.ParseCommand(output)

		// print the output if it is a regular output
		if !isCommand {
			continue
		}

		err := c.processWorkflowCommands(command)
		if err != nil {
			log.Errorf("failed to process workflow command", "error", err.Error(), "output", output)
		}
	}

	return processEnvironmentFiles(ctx, c.stepID, c.envFiles, c.ec)
}

func (c *ContainerExecutor) loadEnvFiles() error {
	dir := config.Client().
		Directory().
		WithNewFile("env", "").
		WithNewFile("path", "").
		WithNewFile("outputs", "").
		WithNewFile("step_summary", "")

	c.envFiles = &EnvironmentFiles{
		Env:         core.NewDaggerEnvironmentFile(dir.File("env")),
		Path:        core.NewDaggerEnvironmentFile(dir.File("path")),
		Outputs:     core.NewDaggerEnvironmentFile(dir.File("outputs")),
		StepSummary: core.NewDaggerEnvironmentFile(dir.File("step_summary")),
	}

	root := filepath.Join(os.Getenv("RUNNER_TEMP"), "env_files")
	env := filepath.Join(root, "env")
	path := filepath.Join(root, "path")
	outputs := filepath.Join(root, "outputs")
	stepSummary := filepath.Join(root, "step_summary")

	c.container = c.container.
		WithMountedDirectory(root, dir).
		WithEnvVariable(core.EnvFileNameGithubEnv, env).
		WithEnvVariable(core.EnvFileNameGithubPath, path).
		WithEnvVariable(core.EnvFileNameGithubActionOutput, outputs).
		WithEnvVariable(core.EnvFileNameGithubStepSummary, stepSummary)

	// update the expression context with the environment files
	c.ec.WithGithubEnv(env).WithGithubPath(path)

	return nil
}

// unloadEnvFiles removes the environment files from the expression context
func (c *ContainerExecutor) unloadEnvFiles() {
	if c.envFiles == nil {
		return
	}

	c.ec.WithoutGithubEnv().WithoutGithubPath()
}

func (c *ContainerExecutor) processWorkflowCommands(cmd *core.WorkflowCommand) error {
	switch cmd.Name {
	case "set-env":
		if err := os.Setenv(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
	case "set-output":
		c.ec.SetStepOutput(c.stepID, cmd.Parameters["name"], cmd.Value)
	case "save-state":
		c.ec.SetStepState(c.stepID, cmd.Parameters["name"], cmd.Value)
	case "add-mask":
		log.Info(fmt.Sprintf("[add-mask] %s", cmd.Value))
	case "add-matcher":
		log.Info(fmt.Sprintf("[add-matcher] %s", cmd.Value))
	case "add-path":
		path := os.Getenv("PATH")
		path = fmt.Sprintf("%s:%s", path, cmd.Value)
		if err := os.Setenv("PATH", path); err != nil {
			return err
		}
	}

	// add the command to the list of commands to keep it as artifact
	c.commands = append(c.commands, cmd)

	return nil
}
