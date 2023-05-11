package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
	"github.com/aweris/gale/model"
	"github.com/aweris/gale/runner/container"
	"github.com/aweris/gale/runner/state"
	"github.com/aweris/gale/runner/steps"
)

type Handler struct {
	state *state.JobRunState
	ch    *container.Handler
	shs   map[string]*steps.Handler
}

func NewHandler(jrs *state.JobRunState) *Handler {
	return &Handler{
		state: jrs,
		ch:    container.NewHandler(jrs.ToContainerState()),
		shs:   make(map[string]*steps.Handler),
	}
}

// TODO: simplify, it's too stupid to have a separate handler for this

func (h *Handler) RunJob(ctx context.Context) error {
	// TODO: maybe we should move this to state.
	// TODO: it's stupid to have a separate runner image per repo. We removed the customizations anyway.
	path, _ := config.SearchDataFile(
		filepath.Join(
			strings.TrimPrefix(h.state.Repo.URL, "https://"),
			"images",
			config.DefaultRunnerLabel,
			config.DefaultRunnerImageTar,
		),
	)

	if path != "" {
		if err := h.ch.LoadContainer(ctx, path); err != nil {
			return err
		}
	} else {
		if err := h.ch.BuildContainer(ctx); err != nil {
			return err
		}
	}

	if err := h.SetupJob(ctx); err != nil {
		return err
	}

	// Run stages
	for _, step := range h.state.Job.Steps {
		h.shs[step.ID].ExecStep(ctx, model.ActionStagePre)
	}

	for _, step := range h.state.Job.Steps {
		h.shs[step.ID].ExecStep(ctx, model.ActionStageMain)
	}

	for _, step := range h.state.Job.Steps {
		h.shs[step.ID].ExecStep(ctx, model.ActionStagePost)
	}

	return nil
}

func (h *Handler) SetupJob(ctx context.Context) error {
	jrc, err := h.state.GetJobRunContext(ctx)
	if err != nil {
		return err
	}

	h.ch.WithExec(ctx, []string{"mkdir", "-p", jrc.Github.Workspace})
	h.ch.WithWorkDir(jrc.Github.Workspace)

	h.ch.WithEnv(jrc.Github.ToEnv())
	h.ch.WithEnv(jrc.Runner.ToEnv())

	// TODO: workaround for now, we should have a better way to do this
	if data, err := json.Marshal(jrc.Github.Event); err == nil {
		h.ch.WithNewFile(
			ctx,
			jrc.Github.EventPath, dagger.ContainerWithNewFileOpts{
				Contents:    string(data),
				Permissions: 0644,
				Owner:       "runner:runner",
			},
		)
	}

	jre := h.state.GetJobRunEnv()

	h.ch.WithEnv(jre.WorkflowEnv)
	h.ch.WithEnv(jre.JobEnv)

	for idx, step := range h.state.Job.Steps {
		if step.ID == "" {
			step.ID = strconv.Itoa(idx)
		}

		h.shs[step.ID] = steps.NewHandler(h.ch, h.state.GetStepRunState(step))

		if step.Uses != "" {
			fmt.Printf("Download action repository '%s'", step.Uses)

			h.shs[step.ID].WithAction(ctx, step.Uses)
		}
	}

	return nil
}
