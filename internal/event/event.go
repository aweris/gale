package event

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"
)

// Event represents a significant change or action that occurs within the system.
type Event[T Context] interface {
	Handle(ctx context.Context, ec *T, publisher Publisher[T]) Result[T]
}

// Status represents the current state of an event.
type Status string

const (
	StatusInProgress = "in_progress"
	StatusSucceeded  = "succeeded"
	StatusSkipped    = "skipped"
	StatusFailed     = "failed"
)

// Result represents the result of an event.
type Result[T Context] struct {
	Status Status `json:"status"`
	Err    error  `json:"error,omitempty"`
	Stdout string `json:"stdout,omitempty"`
}

// Record wraps an event and its result with additional metadata.
type Record[T Context] struct {
	Event[T]
	Result[T]

	ID        int          `json:"id"`
	EventName string       `json:"event_name"`
	Timestamp time.Time    `json:"timestamp"`
	Parent    *Record[T]   `json:"-"` // not serialized to avoid circular references
	Children  []*Record[T] `json:"children,omitempty"`
}

// Context is a marker interface for event contexts to pass to event handlers.
type Context interface {
	// intended to be empty
}

// Publisher is an interface for publishing events.
type Publisher[T Context] interface {
	// Publish publishes an event and returns a record of the event.
	Publish(ctx context.Context, event Event[T]) Record[T]
	publish(ctx context.Context, event Event[T], parent *Record[T]) Record[T]
}

var _ Publisher[Context] = new(StdPublisher[Context])

// StdPublisher is a simple implementation of Publisher.
type StdPublisher[T Context] struct {
	events  []*Record[T]
	counter *atomic.Uint64
	context *T

	// idxRootEvents is an index of the root events in the events slice. This is used to avoid having to iterate
	// over the entire slice to find the root events.
	idxRootEvents []*Record[T]
}

func NewStdPublisher[T Context](context *T) *StdPublisher[T] {
	return &StdPublisher[T]{
		counter: new(atomic.Uint64),
		context: context,
	}
}

func (s *StdPublisher[T]) Publish(ctx context.Context, event Event[T]) Record[T] {
	return s.publish(ctx, event, nil)
}

func (s *StdPublisher[T]) publish(ctx context.Context, event Event[T], parent *Record[T]) Record[T] {
	record := Record[T]{
		ID:        int(s.counter.Add(1)),
		EventName: reflect.TypeOf(event).Name(),
		Event:     event,
		Result:    Result[T]{Status: StatusInProgress},
		Timestamp: time.Now(),
		Parent:    parent,
	}

	// If the parent is not nil, add the record to the parent's children. Otherwise, add it to the root events.
	if parent != nil {
		parent.Children = append(parent.Children, &record)
	} else {
		s.idxRootEvents = append(s.idxRootEvents, &record)
	}

	s.events = append(s.events, &record)

	record.Result = event.Handle(ctx, s.context, &nestedPublisher[T]{parent: &record, StdPublisher: s})

	return record
}

// nestedPublisher is a publisher that wraps another publisher and passes a parent record. This is used to
// create a tree of events without having to pass the parent record to every event handler.
//
// It is not intended to be used directly.
type nestedPublisher[T Context] struct {
	*StdPublisher[T]

	parent *Record[T]
}

func (p *nestedPublisher[T]) Publish(ctx context.Context, event Event[T]) Record[T] {
	return p.publish(ctx, event, p.parent)
}
