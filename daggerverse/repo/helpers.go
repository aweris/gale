package main

// gitContainer returns a container with the git image and the given source mounted at workdir.
func gitContainer(source *Directory) *Container {
	return dag.Container().From("alpine/git:latest").WithMountedDirectory("/src", source).WithWorkdir("/src")
}
