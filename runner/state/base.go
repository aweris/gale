package state

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/model"
	"github.com/aweris/gale/repository"
)

type BaseState struct {
	Repo   *repository.Repo
	Client *dagger.Client
}

func (s *BaseState) DataHome() string {
	return s.Repo.DataHome
}

func (s *BaseState) DataPath(parts ...string) string {
	return s.Repo.DataPath(parts...)
}

func (s *BaseState) GetWorkflow(name string) (*model.Workflow, error) {
	workflow, ok := s.Repo.Workflows[name]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", name)
	}

	return workflow, nil
}
