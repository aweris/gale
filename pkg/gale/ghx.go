package gale

import "dagger.io/dagger"

// TODO: use releases instead of latest. This is temporary hack.

func buildExecutorFromSource(client *dagger.Client) *dagger.File {
	return client.Container().
		From("golang:1.20-bullseye").
		WithExec([]string{"go", "install", "github.com/aweris/ghx@v0.0.1"}).
		Directory("/go/bin").
		File("ghx")
}
