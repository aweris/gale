package preflight

import (
	"context"
	"fmt"
)

// Validator entrypoint for preflight checks.
type Validator struct {
	ctx      *Context
	tasks    map[string]Task
	reporter Reporter
}

// NewValidator creates a new Validator.
func NewValidator(ctx context.Context, reporter Reporter) *Validator {
	return &Validator{
		tasks:    make(map[string]Task),
		ctx:      &Context{Context: ctx},
		reporter: reporter,
	}
}

// Register registers a task to the validator.
func (v *Validator) Register(tasks ...Task) error {
	for _, t := range tasks {
		if _, exists := v.tasks[t.Name()]; exists {
			return fmt.Errorf("task already exists with name %s", t.Name())
		}

		v.tasks[t.Name()] = t
	}

	return nil
}

// Validate validates the preflight checks with given options. Options are optional and only first one is used if multiple
// options are provided.
func (v *Validator) Validate(opts ...Options) error {
	o := Options{}

	if len(opts) > 0 {
		o = opts[0]
	}

	executed := make(map[string]bool)

	for _, t := range v.tasks {
		if err := v.executeTask(t, o, executed); err != nil {
			v.reporter.Report(t, errToResult(err))
			return err
		}
	}

	return nil
}

// executeTask executes a task and its dependencies.
func (v *Validator) executeTask(t Task, opts Options, executed map[string]bool) error {
	// If the task is already executed, skip it.
	if _, ok := executed[t.Name()]; ok {
		return nil
	}

	// Execute the dependencies first.
	for _, dep := range t.DependsOn() {
		// check if the dependency exists
		task, exist := v.tasks[dep]
		if !exist {
			return fmt.Errorf("dependency %s not found for task %s", dep, t.Name())
		}

		// execute the dependency
		if err := v.executeTask(task, opts, executed); err != nil {
			return err
		}
	}

	// Execute the task itself and report the result.
	v.reporter.Report(t, t.Run(v.ctx, opts))

	executed[t.Name()] = true

	return nil
}

func errToResult(err error) Result {
	return Result{Status: Failed, Messages: []Message{{Level: Error, Content: err.Error()}}}
}
