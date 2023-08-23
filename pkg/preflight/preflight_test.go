package preflight_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aweris/gale/pkg/preflight"
)

var _ preflight.Task = new(MockTask)

// Mock implementations for Setup, Check, and Reporter
type MockTask struct {
	name       string
	dependsOn  []string
	runInvoked bool
}

func (m *MockTask) Name() string             { return m.name }
func (m *MockTask) DependsOn() []string      { return m.dependsOn }
func (m *MockTask) Type() preflight.TaskType { return preflight.TaskTypeCheck }
func (m *MockTask) Run(_ *preflight.Context, _ preflight.Options) preflight.Result {
	m.runInvoked = true
	return preflight.Result{Status: preflight.Passed}
}

type MockReporter struct {
	reports []string
}

func (m *MockReporter) Report(t preflight.Task, _ preflight.Result) error {
	m.reports = append(m.reports, t.Name())
	return nil
}

func TestValidator(t *testing.T) {
	// Create the validator with a mock reporter
	reporter := &MockReporter{}
	validator := preflight.NewValidator(reporter)

	// Register setups and checks
	setupA := &MockTask{name: "setupA"}
	checkA := &MockTask{name: "checkA", dependsOn: []string{"setupA"}}
	checkB := &MockTask{name: "checkB", dependsOn: []string{"checkA"}}

	assert.NoError(t, validator.Register(setupA))
	assert.NoError(t, validator.Register(checkA))
	assert.NoError(t, validator.Register(checkB))

	// Execute the validator
	opts := preflight.Options{}
	assert.NoError(t, validator.Validate(opts))

	// Ensure that the setups and checks were invoked in the correct order
	assert.True(t, setupA.runInvoked)
	assert.True(t, checkA.runInvoked)
	assert.True(t, checkB.runInvoked)
	assert.Equal(t, []string{"setupA", "checkA", "checkB"}, reporter.reports)
}
