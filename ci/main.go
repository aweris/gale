package main

type GaleCi struct{}

func (m *GaleCi) Daggerverse() *Daggerverse {
	return &Daggerverse{}
}

func (m *GaleCi) Gha() *Gha {
	return &Gha{}
}
