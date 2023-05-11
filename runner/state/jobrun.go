package state

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aweris/gale/github/cli"
	"github.com/aweris/gale/model"
)

type JobRunState struct {
	BaseState

	parent *WorkflowRunState

	Job *model.Job

	children map[string]*StepRunState
}

func (s *JobRunState) ToContainerState() *ContainerState {
	return &ContainerState{
		BaseState: s.parent.BaseState,
	}
}

func (s *JobRunState) DataHome() string {
	return s.parent.Repo.DataPath("runs", s.parent.WorkflowRun.ID)
}

func (s *JobRunState) DataPath(parts ...string) string {
	return s.parent.Repo.DataPath(append([]string{"runs", s.parent.WorkflowRun.ID}, parts...)...)
}

func (s *JobRunState) GetStepRunState(step *model.Step) *StepRunState {
	if state, ok := s.children[step.ID]; ok {
		return state
	}

	state := &StepRunState{
		BaseState: s.BaseState,
		parent:    s,
		Step:      step,
	}

	s.children[step.ID] = state

	return state
}

func (s *JobRunState) GetJobRunEnv() *model.JobRunEnv {
	return &model.JobRunEnv{
		WorkflowEnv: s.parent.Workflow.Environment,
		JobEnv:      s.Job.Environment,
	}
}

func (s *JobRunState) GetJobRunContext(ctx context.Context) (*model.JobRunContext, error) {
	user, err := cli.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	token, err := cli.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	return &model.JobRunContext{
		Github: &model.GithubContext{
			Actor:             user.Login,
			ActorID:           strconv.Itoa(user.ID),
			ApiURL:            "https://api.github.com",                 // TODO: make this configurable for github enterprise
			Event:             make(map[string]interface{}),             // TODO: generate event data
			EventName:         "push",                                   // TODO: make this configurable, this is for testing purposes
			EventPath:         "/home/runner/_temp/workflow/event.json", // TODO: make this configurable or get from runner
			GraphqlURL:        "https://api.github.com/graphql",         // TODO: make this configurable for github enterprise
			Repository:        s.Repo.NameWithOwner,
			RepositoryID:      s.Repo.ID,
			RepositoryOwner:   s.Repo.Owner.Login,
			RepositoryOwnerID: s.Repo.Owner.ID,
			RepositoryURL:     s.Repo.URL,
			RetentionDays:     0,
			RunID:             "1",
			RunNumber:         "1",
			RunAttempt:        "1",
			SecretSource:      "None",               // TODO: double check if it's possible to get this value from github cli
			ServerURL:         "https://github.com", // TODO: make this configurable for github enterprise
			Token:             token,
			TriggeringActor:   user.Login,
			Workflow:          s.parent.WorkflowRun.Name,
			WorkflowRef:       "", // TODO: fill this value
			WorkflowSHA:       "", // TODO: fill this value
			Workspace:         fmt.Sprintf("/home/runner/work/%s/%s", s.Repo.Name, s.Repo.Name),
		},
		Runner: &model.RunnerContext{
			Name:      "", // TODO: Not sure if we need this at all. Remove after double-checking.
			OS:        "linux",
			Arch:      "x64", // TODO: This should be determined by the host
			Temp:      "/home/runner/_temp",
			ToolCache: "/home/runner/_tool",
			Debug:     "", // TODO: "1" for debug mode, "" for normal mode get it from config
		},
	}, nil
}
