package executor

import (
	"context"

	runnerpkg "github.com/aweris/gale/runner"
)

type StepExecutor interface {
	pre(ctx context.Context, runner *runnerpkg.Runner) error
	main(ctx context.Context, runner *runnerpkg.Runner) error
	post(ctx context.Context, runner *runnerpkg.Runner) error
}
