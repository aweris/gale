package main

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct{}

func (g *Gale) Workflows() *Workflows {
	return new(Workflows)
}
