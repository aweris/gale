package main

type ActionsGenerator struct{}

// Generates a dagger module from based on the given Github Actions repository.
func (m *ActionsGenerator) Generate(
	// The Github Actions repository to generate dagger modules for. Format: <action-repo>@<version>
	action string,
	// Version of the aweris/gale to use.
	// +optional=true
	// +default=f68252f64186ad05c1c8b3d72a25e7c2933c169d
	galeVersion string,
	// Version of the dagger to use.
	// +optional=true
	daggerVersion string,
) (*Directory, error) {
	ca, err := loadCustomAction(action)
	if err != nil {
		return nil, err
	}

	// create a new dagger module
	module := NewDaggerCli(daggerVersion).InitModule(ca.RepoName, "github.com/aweris/gale@"+galeVersion)

	// replace the initial main.go with the generated main.go
	module = module.WithoutFile("main.go").WithFile("main.go", generate(ca))

	// add README.md
	module = module.WithFile("README.md", generateModuleREADME(ca, daggerVersion))

	// add .gitignore file to ignore generated files
	module = module.WithNewFile(".gitignore", "/dagger.gen.go\n/internal/querybuilder/\n/querybuilder/\n")

	return dag.Directory().WithDirectory(ca.Repo, module, DirectoryWithDirectoryOpts{Include: []string{"**/*", ".git*"}}), nil
}
