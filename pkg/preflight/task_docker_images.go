package preflight

import (
	"fmt"

	"github.com/aweris/gale/internal/config"
)

var _ Task = new(DockerImagesCheck)

// DockerImagesCheck is a preflight check that checks if all docker images are available.
type DockerImagesCheck struct{}

func (d *DockerImagesCheck) Name() string {
	return NameDockerImagesCheck
}

func (d *DockerImagesCheck) Type() TaskType {
	return TaskTypeCheck
}

func (d *DockerImagesCheck) DependsOn() []string {
	return []string{NameDaggerCheck, NameWorkflowLoader}
}

func (d *DockerImagesCheck) Run(ctx *Context, _ Options) Result {
	var (
		status = Passed
		msg    = make([]Message, 0)
	)
	// validate runner image
	_, err := config.Client().Container().From(config.RunnerImage()).Sync(ctx.Context)
	if err != nil {
		status = Failed
		msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Runner image %s is not available", config.RunnerImage())})
	} else {
		msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Runner image %s is available", config.RunnerImage())})
	}

	for image := range ctx.DockerImages {
		_, err := config.Client().Container().From(image).Sync(ctx.Context)
		if err != nil {
			status = Failed
			msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Docker image %s is not available", image)})
		} else {
			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Docker image %s is available", image)})
		}
	}

	if status == Passed {
		msg = append(msg, Message{Level: Info, Content: "All docker images are available"})
	}

	return Result{Status: status, Messages: msg}
}
