package services

import "dagger/gale/internal/dagger"

type Services struct {
	dag *dagger.Client
}

func NewServices(dag *dagger.Client) *Services {
	return &Services{dag: dag}
}
