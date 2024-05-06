package services

import (
	"context"
	"fmt"

	"dagger/gale/internal/dagger"
	"dagger/gale/tools/gobuild"
)

func (s *Services) WithGHX(ctx context.Context, ctr *dagger.Container) (*dagger.Container, error) {
	binary, err := gobuild.Build(
		ctx,
		s.dag,
		gobuild.Workdir("services/ghx"),
		gobuild.Include("services/ghx/**/*.go", "services/ghx/go.*", "libs/common/**/*.go", "libs/common/go.*"),
	)
	if err != nil {
		return nil, err
	}

	ctr = ctr.WithFile("/usr/local/bin/ghx", binary, dagger.ContainerWithFileOpts{Permissions: 0777})

	path, err := ctr.EnvVariable(ctx, "PATH")
	if err != nil {
		return nil, err
	}

	return ctr.WithEnvVariable("PATH", fmt.Sprintf("%s:/usr/local/bin", path)), nil
}
