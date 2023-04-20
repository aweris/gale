package runner

import (
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"

	"github.com/google/uuid"

	"github.com/aweris/gale/gha"
)

// WithEnvironment adds the given environment variables to the container.
func (r *Runner) WithEnvironment(env gha.Environment) {
	for k, v := range env {
		r.Container = r.Container.WithEnvVariable(k, v)
	}
}

// WithoutEnvironment removes given environment variables from the container. If a fallback environment is given,
// instead of removing the variable, it will be set to the value of the fallback environment.
//
// If multiple fallback environments are given, they will be merged in the order they are given. The last environment
// in the list will have the highest priority.
//
// This is useful for removing overridden environment variables without losing the original value.
//
// Example:
//
//	runner.WithEnvironment(gha.Environment{ "FOO": "bar"})
//	runner.WithoutEnvironment(gha.Environment{"FOO": "bar"}, gha.Environment{"FOO": "qux"})
//
// The above example will result in the environment variable FOO being set to "qux" instead of being removed.
func (r *Runner) WithoutEnvironment(env gha.Environment, fallback ...gha.Environment) {
	merged := gha.Environment{}

	for _, environment := range fallback {
		// to merge the fallback environments with priority, we need to merge them in order.
		// the last environment in the list will have the highest priority.
		merged = merged.Merge(environment)
	}

	for k := range env {
		if _, ok := merged[k]; ok {
			r.Container = r.Container.WithEnvVariable(k, merged[k])
		} else {
			r.Container = r.Container.WithoutEnvVariable(k)
		}
	}
}

// WithInputs transform given input name as INPUT_<NAME> and add it to the container as environment variable.
func (r *Runner) WithInputs(inputs map[string]string) {
	for k, v := range inputs {
		// TODO: This is a hack to get around the fact that we can't set the GITHUB_TOKEN as an input. Remove this
		// once we have a better solution.
		if strings.TrimSpace(v) == "${{ secrets.GITHUB_TOKEN }}" {
			v = os.Getenv("GITHUB_TOKEN")
		}
		r.Container = r.Container.WithEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)), v)
	}
}

// WithoutInputs removes the given inputs from the container.
func (r *Runner) WithoutInputs(inputs map[string]string) {
	for k := range inputs {
		r.Container = r.Container.WithoutEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)))
	}
}

// WithTempDirectory adds the given directory as /home/runner/_temp/<UUID> and returns the path to the directory.
func (r *Runner) WithTempDirectory(dir *dagger.Directory) string {
	path := fmt.Sprintf("/home/runner/_temp/%s", uuid.New())

	r.Container = r.Container.WithDirectory(path, dir)

	return path
}

// WithExec is simple wrapper around dagger.Container.WithExec. This is useful for simplifying the syntax when
// using this method.
func (r *Runner) WithExec(cmd string, args ...string) {
	r.Container = r.Container.WithExec(append([]string{cmd}, args...))
}
