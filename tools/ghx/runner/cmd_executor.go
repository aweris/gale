package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/tools/ghx/actions"
	"github.com/aweris/gale/tools/ghx/log"
)

type CmdExecutor struct {
	args     []string                // args to pass to the command
	env      map[string]string       // env to pass to the command as environment variables
	ec       *actions.ExprContext    // ec is the expression context to evaluate the github expressions
	commands []*core.WorkflowCommand // commands is the list of commands that are executed in the step
	envFiles *EnvironmentFiles       // envFiles contains temporary files that can be used to perform certain actions

}

type EnvironmentFiles struct {
	Env         *core.EnvironmentFile // Env is the environment file that holds the environment variables
	Path        *core.EnvironmentFile // Path is the environment file that holds the path variables
	Outputs     *core.EnvironmentFile // Outputs is the environment file that holds the outputs
	StepSummary *core.EnvironmentFile // StepSummary is the environment file that holds the step summary
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
		ec:   sa.runner.context,
	}
}

func NewCmdExecutorFromStepRun(sr *StepRun) *CmdExecutor {
	args := []string{sr.Shell}

	args = append(args, sr.ShellArgs...)
	args = append(args, sr.Path)

	return &CmdExecutor{
		args: args,
		env:  make(map[string]string),
		ec:   sr.runner.context,
	}
}

func (c *CmdExecutor) Execute(_ context.Context) error {
	cmd := exec.Command(c.args[0], c.args[1:]...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	rawout := bytes.NewBuffer(nil)

	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)

	// load environment files - this will create env files and load it to the environment. That's why we need to do this
	// before setting the environment variables
	err := c.loadEnvFiles()
	if err != nil {
		return err
	}
	defer c.unloadEnvFiles()

	// add environment variables

	env := os.Environ()

	for k, v := range c.env {
		// convert value to Evaluable String type
		str := actions.NewString(v)

		// evaluate the expression
		res, err := str.Eval(c.ec)
		if err != nil {
			log.Errorf("failed to evaluate value", "error", err.Error(), "key", k, "value", v)

			return fmt.Errorf("failed to evaluate default value for input %s: %v", k, err)
		}

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

			// write to stdout as it is so we can keep original formatting
			rawout.WriteString(output)
			rawout.WriteString("\n") // scanner strips newlines

			isCommand, command := core.ParseCommand(output)

			// print the output if it is a regular output
			if !isCommand {
				log.Info(output)

				// write to stdout as it is so we can keep original formatting
				stdout.WriteString(output)
				stdout.WriteString("\n") // scanner strips newlines

				continue
			}

			// add the command to the list of commands to keep it as artifact
			c.commands = append(c.commands, command)
		}
	}()

	return cmd.Wait()
}

func (c *CmdExecutor) loadEnvFiles() error {
	if c.envFiles == nil {
		c.envFiles = &EnvironmentFiles{}
	}

	// TODO: move this to a better place. No need read os env directly here
	dir, err := os.MkdirTemp(os.Getenv("RUNNER_TEMP"), "env_files")
	if err != nil {
		return err
	}

	env, err := core.NewEnvironmentFile(filepath.Join(dir, "env"))
	if err != nil {
		return err
	}

	c.env[core.EnvFileNameGithubEnv] = env.Path
	c.envFiles.Env = env

	path, err := core.NewEnvironmentFile(filepath.Join(dir, "path"))
	if err != nil {
		return err
	}

	c.env[core.EnvFileNameGithubPath] = path.Path
	c.envFiles.Path = path

	outputs, err := core.NewEnvironmentFile(filepath.Join(dir, "outputs"))
	if err != nil {
		return err
	}

	c.env[core.EnvFileNameGithubActionOutput] = outputs.Path
	c.envFiles.Outputs = outputs

	stepSummary, err := core.NewEnvironmentFile(filepath.Join(dir, "step_summary"))
	if err != nil {
		return err
	}

	c.env[core.EnvFileNameGithubStepSummary] = stepSummary.Path
	c.envFiles.StepSummary = stepSummary

	// update the expression context with the environment files
	c.ec.WithGithubEnv(env).WithGithubPath(path)

	return nil
}

// unloadEnvFiles removes the environment files from the expression context
func (c *CmdExecutor) unloadEnvFiles() {
	if c.envFiles == nil {
		return
	}

	c.ec.WithoutGithubEnv().WithoutGithubPath()
}
