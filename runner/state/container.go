package state

import "dagger.io/dagger"

type ContainerState struct {
	BaseState

	Container *dagger.Container
}
