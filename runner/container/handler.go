package container

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/builder"
	"github.com/aweris/gale/runner/state"
	"github.com/google/uuid"
)

type Handler struct {
	state *state.ContainerState
}

func NewHandler(state *state.ContainerState) *Handler {
	return &Handler{state: state}
}

func (h *Handler) BuildContainer(ctx context.Context) error {
	container, err := builder.NewBuilder(h.state.Client, h.state.Repo).Build(ctx)
	if err != nil {
		return err
	}

	h.state.Container = container

	return nil
}

func (h *Handler) LoadContainer(_ context.Context, path string) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	h.state.Container = h.state.Client.Container().Import(h.state.Client.Host().Directory(dir).File(base))

	return nil
}

// WithExecOpts is the optional configuration for WithExec.
type WithExecOpts struct {
	Execute bool                         // Execute is a flag to indicate whether the command should be executed immediately.
	Strace  bool                         // Strace is a flag to indicate whether the command should be executed with strace.
	Opts    dagger.ContainerWithExecOpts // Extra dagger options to be passed to the container.
}

// WithExecResult is the result of WithExec. Only populated when Execute is true.
type WithExecResult struct {
	Stdout string // Stdout is the stdout of the command.
	Strace string // Strace is the strace output of the command.
}

func (h *Handler) WithExec(ctx context.Context, args []string, opts ...WithExecOpts) (*WithExecResult, error) {
	opt := WithExecOpts{}

	if len(opts) > 0 {
		opt = opts[0]
	}

	var straceLogPath = fmt.Sprintf("/tmp/strace-%s.log", uuid.New())

	if opt.Execute && opt.Strace {
		args = append([]string{"strace", "-o", straceLogPath}, args...)
	}

	h.state.Container = h.state.Container.WithExec(args, opt.Opts)

	// if execute is false, return empty result. Caller shouldn't be interested in the result since it is not executed.
	if !opt.Execute {
		return &WithExecResult{}, nil
	}

	out, err := h.state.Container.Stdout(ctx)
	if err != nil {
		return &WithExecResult{Stdout: out}, err
	}

	strace := ""

	if opt.Strace {
		contents, err := h.state.Container.File(straceLogPath).Contents(ctx)
		if err != nil {
			return &WithExecResult{Stdout: out}, err
		}

		strace = contents
	}

	return &WithExecResult{Stdout: out, Strace: strace}, nil
}

func (h *Handler) WithNewFile(_ context.Context, path string, opts ...dagger.ContainerWithNewFileOpts) {
	opt := dagger.ContainerWithNewFileOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	h.state.Container = h.state.Container.WithNewFile(path, opt)
}

func (h *Handler) WithDirectory(_ context.Context, path string, directory *dagger.Directory, opts ...dagger.ContainerWithDirectoryOpts) {
	opt := dagger.ContainerWithDirectoryOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	h.state.Container = h.state.Container.WithDirectory(path, directory, opt)
}

func (h *Handler) WithWorkDir(path string) {
	h.state.Container = h.state.Container.WithWorkdir(path)
}

func (h *Handler) WithPath(ctx context.Context, newPath string) error {
	existingPath, err := h.state.Container.EnvVariable(ctx, "PATH")
	if err != nil {
		return fmt.Errorf("failed to get PATH: %w", err)
	}

	h.state.Container = h.state.Container.WithEnvVariable("PATH", fmt.Sprintf("%s:%s", existingPath, newPath))

	return nil
}

func (h *Handler) WithEnv(env map[string]string) {
	for k, v := range env {
		h.WithEnvVariable(k, v)
	}
}

func (h *Handler) WithoutEnv(env map[string]string, fallbacks ...map[string]string) {
	merged := map[string]string{}

	for _, environment := range fallbacks {
		for k, v := range environment {
			merged[k] = v
		}
	}

	for k := range env {
		if _, ok := merged[k]; ok {
			h.WithEnvVariable(k, merged[k])
		} else {
			h.WithoutEnvVariable(k)
		}
	}
}

func (h *Handler) WithEnvVariable(name, value string) {
	h.state.Container = h.state.Container.WithEnvVariable(name, value)
}

func (h *Handler) WithoutEnvVariable(name string) {
	h.state.Container = h.state.Container.WithoutEnvVariable(name)
}
