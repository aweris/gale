package state

import (
	"fmt"
	"time"

	"github.com/aweris/gale/model"
)

type WorkflowRunState struct {
	BaseState

	Workflow    *model.Workflow
	WorkflowRun *model.WorkflowRun

	children map[string]*JobRunState
}

func NewWorkflowRunState(base *BaseState) *WorkflowRunState {
	return &WorkflowRunState{
		BaseState: *base,
		children:  make(map[string]*JobRunState),
	}
}

func (s *WorkflowRunState) NewWorkflowRun(name string) error {
	workflow, err := s.GetWorkflow(name)
	if err != nil {
		return err
	}

	s.Workflow = workflow
	s.WorkflowRun = &model.WorkflowRun{
		ID:   time.Now().Format(time.RFC3339Nano), // TODO: generate a real ID, just a hack for now
		Name: name,
	}

	return nil
}

func (s *WorkflowRunState) RunStart() {
	s.WorkflowRun.RunStartedAt = time.Now()
}

func (s *WorkflowRunState) RunEnd() {
	s.WorkflowRun.RunDuration = time.Since(s.WorkflowRun.RunStartedAt)
}

func (s *WorkflowRunState) UpdateStatus(status model.WorkflowRunStatus) {
	s.WorkflowRun.Status = status
}

func (s *WorkflowRunState) GetJobs() map[string]*model.Job {
	return s.Workflow.Jobs
}

func (s *WorkflowRunState) GetJob(name string) (*model.Job, error) {
	workflow, err := s.GetWorkflow(s.WorkflowRun.Name)
	if err != nil {
		return nil, err
	}

	job, ok := workflow.Jobs[name]
	if !ok {
		return nil, fmt.Errorf("job %s/%s not found", s.WorkflowRun.Name, job.Name)
	}

	if job.Name == "" {
		job.Name = name
	}

	return job, nil
}

func (s *WorkflowRunState) GetJobRunState(job *model.Job) *JobRunState {
	if state, ok := s.children[job.Name]; ok {
		return state
	}

	state := &JobRunState{
		BaseState: s.BaseState,
		parent:    s,
		Job:       job,
		children:  make(map[string]*StepRunState),
	}

	s.children[job.Name] = state

	return state
}
