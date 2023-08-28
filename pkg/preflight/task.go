package preflight

// TaskType is the type of the task. It can be either check or load.
type TaskType string

const (
	TaskTypeCheck TaskType = "check"
	TaskTypeLoad  TaskType = "load"
)

// Task is a task that can be executed.
type Task interface {
	Name() string                          // Name returns the name of the task.
	Type() TaskType                        // Type returns the type of the task.
	DependsOn() []string                   // DependsOn returns the list of tasks that this task depends on.
	Run(ctx *Context, opts Options) Result // Run runs the task and returns the result.
}

// Task names to make it easier to reference them in dependencies.
const (
	NameDaggerCheck       = "Dagger"
	NameGHCheck           = "GitHub CLI"
	NameRepoLoader        = "Repo"
	NameWorkflowLoader    = "Workflow"
	NameDockerImagesCheck = "Docker Images"
	NameStepShellCheck    = "Run Step Shell"
)

// StandardTasks returns the standard tasks that are used in preflight checks.
func StandardTasks() []Task {
	return []Task{
		new(DaggerCheck),
		new(GHCheck),
		new(RepoLoader),
		new(WorkflowLoader),
		new(DockerImagesCheck),
		new(StepShellCheck),
	}
}
