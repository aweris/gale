package main

import "context"

type Ghx struct{}

func (m *Ghx) source() *Source {
	return &Source{}
}

// Binary adds the ghx binary to the given container and adds binary to the PATH environment variable.
func (m *Ghx) Binary(ctx context.Context, container *Container) (*Container, error) {
	return m.source().Binary(ctx, container)
}
