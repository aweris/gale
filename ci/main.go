package main

type GaleCi struct{}

func (m *GaleCi) Daggerverse() *Daggerverse {
	return &Daggerverse{}
}
