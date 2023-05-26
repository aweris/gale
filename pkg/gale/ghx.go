package gale

import "dagger.io/dagger"

// TODO: use releases instead of latest. This is temporary hack.

func buildExecutorFromSource(client *dagger.Client) *dagger.File {
	return client.Container().
		From("golang:1.20-bullseye").
		WithExec([]string{"go", "install", "github.com/aweris/ghx@ba9b6144fa7bb283d56f0ad52e9e174239a7aded"}).
		Directory("/go/bin").
		File("ghx")
}
