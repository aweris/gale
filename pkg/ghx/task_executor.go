package ghx

import (
	"context"
	"fmt"
	"time"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/log"
)

// TaskExecutor is a task executor that runs a task and keeps status, conclusion and timing information about
// the execution.
type TaskExecutor[T TaskResult] struct {
	Name        string               // Name of the execution
	Ran         bool                 // Ran indicates if the execution ran
	Status      core.Status          // Status of the execution
	Result      *T                   // Result of the execution
	StartedAt   time.Time            // StartedAt time of the execution
	CompletedAt time.Time            // CompletedAt time of the execution
	executorFn  TaskExecutorFn[T]    // executorFn is the function to be executed
	conditionFn TaskConditionalFn[T] // conditionFn is the function that determines if the task should be executed
}

// TaskResult is a marker interface for the response of the task executor.
type TaskResult interface {
	// GetConclusion returns the conclusion of the task.
	GetConclusion() core.Conclusion
}

// TaskExecutorFn is the function that will be executed by the task executor.
//
// The return values are:
//   - conclusion: conclusion of the execution
//   - err: error if any
type TaskExecutorFn[T TaskResult] func(ctx context.Context) (result *T, err error)

// TaskConditionalFn is the function that determines if the task should be executed. If the task should not be
// executed, the conclusion is returned as well.
//
// The return values are:
//   - run: indicates if the task should be executed
//   - conclusion: conclusion of the execution if the task should not be executed (run = false)
//   - err: error if any
//
// The rules are:
//   - If err is not nil, run is false and conclusion is ignored.
//   - If run is true, conclusion and err are ignored.
//   - If run is false, conclusion is not empty and err is nil, do not run the task and set the conclusion.
//   - If run is false, conclusion is empty and err is nil, invalid task. Ignore it completely. The reason this
//     scenario exist is that the task is not invalid by itself. It is just not applicable to the current context, and
//     we can't determine if it is invalid or not in planning phase.
type TaskConditionalFn[T TaskResult] func(ctx context.Context) (run bool, result *T, err error)

// NewTaskExecutor creates a new task executor.
func NewTaskExecutor[T TaskResult](name string, fn TaskExecutorFn[T]) TaskExecutor[T] {
	return NewConditionalTaskExecutor(name, fn, nil)
}

// NewConditionalTaskExecutor creates a new task executor.
func NewConditionalTaskExecutor[T TaskResult](name string, executorFn TaskExecutorFn[T], conditionalFn TaskConditionalFn[T]) TaskExecutor[T] {
	return TaskExecutor[T]{
		Name:        name,
		Status:      core.StatusQueued,
		executorFn:  executorFn,
		conditionFn: conditionalFn,
	}
}

// Run runs the task and updates the status, conclusion and timing information.
func (t *TaskExecutor[T]) Run(ctx context.Context) (run bool, result *T, err error) {
	t.StartedAt = time.Now()
	t.Status = core.StatusInProgress

	if t.conditionFn != nil {
		run, result, err = t.conditionFn(ctx)
		if !run {
			t.Ran = run
			t.Result = result
			t.CompletedAt = time.Now()
			t.Status = core.StatusCompleted

			if result != nil {
				// convert pointer type to value type to be able to access interface type methods.
				val := *result

				log.Info(fmt.Sprintf("%s (%s)", t.Name, val.GetConclusion()))
			}

			return run, result, err
		}
	}

	// create ger group for step
	log.Info(t.Name)
	log.StartGroup()
	defer log.EndGroup()

	result, err = t.executorFn(ctx)

	t.Ran = true
	t.Result = result
	t.CompletedAt = time.Now()
	t.Status = core.StatusCompleted

	return run, result, err
}
