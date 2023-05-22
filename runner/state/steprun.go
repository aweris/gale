package state

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aweris/gale/model"
)

type StepRunState struct {
	BaseState

	parent *JobRunState

	Step       *model.Step       // definition of the step
	Result     *model.StepResult // result of the step
	Action     *model.Action     // definition of the action. Only set if the step is an action step
	ActionPath string            // path to the action container. Only set if the step is an action step
	State      map[string]string // state of the step
}

// GetStateEnv returns the state of the step as a map of environment variables formatted as STATE_<key>=<value>
// according to the convention of the GitHub Actions runner.
func (s *StepRunState) GetStateEnv() map[string]string {
	env := make(map[string]string, len(s.State))

	for k, v := range s.State {
		env[fmt.Sprintf("STATE_%s", k)] = v
	}

	return env
}

func (s *StepRunState) GetInputEnv() map[string]string {
	env := make(map[string]string, len(s.Step.With))

	for k, v := range s.Step.With {
		// TODO workaround
		if strings.TrimSpace(v) == "${{ secrets.GITHUB_TOKEN }}" {
			jrc, _ := s.parent.GetJobRunContext(context.Background())
			v = jrc.Github.Token
		}

		env[fmt.Sprintf("INPUT_%s", strings.ToUpper(k))] = v
	}

	return env
}

func (s *StepRunState) SetOutput(name string, value string) {
	if s.Result == nil {
		s.Result = &model.StepResult{}
	}

	if s.Result.Outputs == nil {
		s.Result.Outputs = make(map[string]string)
	}

	s.Result.Outputs[name] = value
}

func (s *StepRunState) SaveState(name string, value string) {
	if s.State == nil {
		s.State = make(map[string]string)
	}

	s.State[name] = value
}

func (s *StepRunState) ExportLogArtifact(name string, content string) error {
	path := s.parent.DataPath("steps", s.Step.ID, "artifacts")

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, name), []byte(content), 0600)
}

func (s *StepRunState) GetStepFallbackEnvs() []map[string]string {
	jre := s.parent.GetJobRunEnv()
	return []map[string]string{jre.WorkflowEnv, jre.JobEnv}
}
