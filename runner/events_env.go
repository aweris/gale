package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/github/actions"
	"github.com/aweris/gale/internal/event"
)

// Environment Events

var (
	_ event.Event[Context] = new(WithEnvironmentEvent)
	_ event.Event[Context] = new(WithoutEnvironmentEvent)
	_ event.Event[Context] = new(AddEnvEvent)
	_ event.Event[Context] = new(ReplaceEnvEvent)
	_ event.Event[Context] = new(RemoveEnvEvent)
)

// WithEnvironmentEvent sets the given environment variables on the container. While setting the environment variables,
// - if the variable is already set, it will publish a ReplaceEnvEvent to replace the value of the variable.
// - if the variable is not set, it will publish a AddEnvEvent to add the variable.
// - if the variable is already set and value is the same, it will do nothing.
type WithEnvironmentEvent struct {
	Env actions.Environment
}

func (e WithEnvironmentEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
	for k, v := range e.Env {
		if val, _ := ec.container.EnvVariable(ctx, k); val != "" {
			if val == v {
				ec.log.Debug(fmt.Sprintf("environment variable %s already set to %s", k, v))
				continue
			}
			publisher.Publish(ctx, ReplaceEnvEvent{Name: k, OldValue: val, NewValue: v})
		} else {
			publisher.Publish(ctx, AddEnvEvent{Name: k, Value: v})
		}
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// WithoutEnvironmentEvent removes given environment variables from the container. If a fallback environment is given,
// instead of removing the variable, it will be set to the value of the fallback environment.
//
// If multiple fallback environments are given, they will be merged in the order they are given. The last environment
// in the list will have the highest priority.
//
// This is useful for removing overridden environment variables without losing the original value.
type WithoutEnvironmentEvent struct {
	Env          actions.Environment
	FallbackEnvs []actions.Environment
}

func (e WithoutEnvironmentEvent) Handle(ctx context.Context, _ *Context, publisher event.Publisher[Context]) event.Result[Context] {
	merged := actions.Environment{}

	for _, environment := range e.FallbackEnvs {
		// to merge the fallback environments with priority, we need to merge them in order.
		// the last environment in the list will have the highest priority.
		merged = merged.Merge(environment)
	}

	for k, v := range e.Env {
		if _, ok := merged[k]; ok {
			publisher.Publish(ctx, ReplaceEnvEvent{Name: k, OldValue: v, NewValue: merged[k]})
		} else {
			publisher.Publish(ctx, RemoveEnvEvent{Name: k})
		}
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// AddEnvEvent introduces new env variable to runner container.
type AddEnvEvent struct {
	Name  string
	Value string
}

func (e AddEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithEnvVariable(e.Name, e.Value)
	return event.Result[Context]{Status: event.StatusSucceeded}
}

// ReplaceEnvEvent replaces existing env Value with the new one. Event assumes existing env and Value validated
// during event creation. Event will not check if the env exists or not.
type ReplaceEnvEvent struct {
	Name     string
	OldValue string // this is used for informational purposes only, not used in event handling.
	NewValue string
}

func (e ReplaceEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithEnvVariable(e.Name, e.NewValue)
	return event.Result[Context]{Status: event.StatusSucceeded}
}

// RemoveEnvEvent removes an env Value from runner container
type RemoveEnvEvent struct {
	Name string
}

func (e RemoveEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithoutEnvVariable(e.Name)
	return event.Result[Context]{Status: event.StatusSucceeded}
}
