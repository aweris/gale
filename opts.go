package main

// FIXME: currently Object types can't be used as a type parameter. So, we can't use WorkflowOpts as a type parameter
//  refactor this code when dagger supports Opt struct types as type parameters.

type WorkflowRunOpts struct {
	// External workflow file to run against the repository.
	WorkflowFile *File

	// Name or path of the workflow to run.
	Workflow string

	// Job name to run. If not specified, all jobs in the workflow will be run.
	Job string
}

type EventOpts struct {
	// Event name.
	Name string

	// File containing the event data in JSON format.
	File *File
}

type RunnerOpts struct {
	// Base container for the runner.
	Ctr *Container

	// Debug flag for the runner.
	Debug bool

	// Enables native docker support to able to run docker commands directly in the workflow.
	UseNativeDocker bool

	// Docker host to use for the runner.
	DockerHost string

	// Enables docker-in-dagger support to be able to run docker commands isolated from the host.
	// Enabling DinD may lead to longer execution times.
	UseDind bool
}

type SecretOpts struct {
	// GitHub token to use for the runner.
	Token *Secret
}
