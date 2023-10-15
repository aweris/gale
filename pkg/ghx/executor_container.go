package ghx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/log"
)

var _ Executor = new(ContainerExecutor)

type ContainerExecutor struct {
	container  *dagger.Container       // container is the container to execute
	entrypoint string                  // entrypoint is the entrypoint of the container
	args       []string                // args is the arguments of the container
	env        map[string]string       // env to pass to the command as environment variables
	dec        *gctx.Context           // dec is the dagger expression context to evaluate action meta files. TODO: temporary solution, we need to find a better way to do this
	commands   []*core.WorkflowCommand // commands is the list of commands that are executed in the step
	envFiles   *EnvironmentFiles       // envFiles contains temporary files that can be used to perform certain actions
}

func NewContainerExecutorFromStepDocker(ctx *gctx.Context, sd *StepDocker) *ContainerExecutor {
	return &ContainerExecutor{
		entrypoint: sd.Step.With["entrypoint"],
		args:       []string{sd.Step.With["args"]},
		env:        sd.Step.Environment,
		dec:        ctx,
		container:  sd.container,
	}
}

func NewContainerExecutorFromStepAction(ctx *gctx.Context, sa *StepAction, entrypoint string) *ContainerExecutor {
	// get context as value to avoid changing the original context. We need to do this because we are going to change
	// inputs context and we don't want to change the original context.
	dec := *ctx

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
	for k, v := range ctx.Steps[sa.Step.ID].State {
		env[fmt.Sprintf("STATE_%s", k)] = v
	}

	// add step level environment variables
	for k, v := range ctx.Env {
		env[k] = v
	}

	var args []string

	args = append(args, sa.Action.Meta.Runs.Args...)

	return &ContainerExecutor{
		entrypoint: entrypoint,
		args:       args,
		env:        env,
		dec:        &dec,
		container:  sa.container,
	}
}

func (c *ContainerExecutor) Execute(ctx *gctx.Context) error {
	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	err := c.loadEnvFiles(ctx)
	if err != nil {
		return err
	}
	defer c.unloadEnvFiles(ctx)

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
		res := str.Eval(c.dec)

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
	out, err := c.container.Stdout(ctx.Context)
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

		err := c.processWorkflowCommands(ctx, command)
		if err != nil {
			log.Errorf("failed to process workflow command", "error", err.Error(), "output", output)
		}
	}

	return processEnvironmentFiles(ctx, c.envFiles, ctx)
}

func (c *ContainerExecutor) loadEnvFiles(ctx *gctx.Context) error {
	dir := ctx.Dagger.Client.
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
		WithEnvVariable(core.EnvFileNameGithubOutput, outputs).
		WithEnvVariable(core.EnvFileNameGithubStepSummary, stepSummary)

	// update the expression context with the environment files
	ctx.WithGithubEnv(env).WithGithubPath(path)

	return nil
}

// unloadEnvFiles removes the environment files from the expression context
func (c *ContainerExecutor) unloadEnvFiles(ctx *gctx.Context) {
	if c.envFiles == nil {
		return
	}

	ctx.WithoutGithubEnv().WithoutGithubPath()
}

func (c *ContainerExecutor) processWorkflowCommands(ctx *gctx.Context, cmd *core.WorkflowCommand) error {
	switch cmd.Name {
	case "set-env":
		if err := os.Setenv(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
	case "set-output":
		ctx.SetStepOutput(cmd.Parameters["name"], cmd.Value)
	case "save-state":
		ctx.SetStepState(cmd.Parameters["name"], cmd.Value)
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
