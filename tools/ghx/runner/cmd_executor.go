package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/tools/ghx/actions"
)

type CmdExecutor struct {
	args     []string                // args to pass to the command
	env      map[string]string       // env to pass to the command as environment variables
	ec       *actions.ExprContext    // ec is the expression context to evaluate the github expressions
	commands []*core.WorkflowCommand // commands is the list of commands that are executed in the step
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
		ec:   actions.NewExprContext(),
	}
}

func NewCmdExecutorFromStepRun(sr *StepRun) *CmdExecutor {
	args := []string{sr.Shell}

	args = append(args, sr.ShellArgs...)
	args = append(args, sr.Path)

	return &CmdExecutor{
		args: args,
		env:  make(map[string]string),
		ec:   actions.NewExprContext(),
	}
}

func (c *CmdExecutor) Execute(_ context.Context) error {
	cmd := exec.Command(c.args[0], c.args[1:]...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	rawout := bytes.NewBuffer(nil)

	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)

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
