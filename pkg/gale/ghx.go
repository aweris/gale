package gale

import "dagger.io/dagger"

// TODO: use releases instead of latest. This is temporary hack.

func buildExecutorFromSource(client *dagger.Client) *dagger.File {
	return client.Container().
		From("golang:1.20-bullseye").
		WithExec([]string{"go", "install", "github.com/aweris/ghx@a9a420e1f5dc2cb4dbdd9f7e3507cbe61a7e5620"}).
		Directory("/go/bin").
		File("ghx")
}
