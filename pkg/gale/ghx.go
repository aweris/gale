package gale

import "dagger.io/dagger"

// TODO: use releases instead of main branch. This is temporary hack.

func buildExecutorFromSource(client *dagger.Client) *dagger.File {
	dir := client.Git("github.com/aweris/ghx").Branch("main").Tree()

	return client.Container().
		From("golang:1.20-bullseye").
		WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("/var/run/docker.sock")).
		WithDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o", "/src/build/ghx", "."}).
		Directory("/src/build").
		File("ghx")
}
