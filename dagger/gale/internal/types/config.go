package types

import "dagger/gale/internal/dagger"

type Config struct {
	Source       *dagger.Directory
	Repo         string
	Tag          string
	Branch       string
	WorkflowsDir string
	Workflow     string
	WorkflowFile *dagger.File
	Job          string
	Event        string
	EventFile    *dagger.File
	Ctr          *dagger.Container
	Token        *dagger.Secret
	Debug        bool
	UseDinD      bool
}
