package main

import (
	"ghx/context"
)

// Executor is the interface that defines contract for objects capable of performing an execution task.
type Executor interface {
	// Execute performs the execution of a specific task with the given context.
	Execute(ctx *context.Context) error
}
