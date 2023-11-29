package main

// default entry point for internal submodules. The intention of this entry point is keep the module clean and
// consistent. This entrypoint is not intended to be used by external modules.
var internal Internal

type Internal struct{}

func (_ *Internal) runner() *Runner {
	return &Runner{}
}

func (_ *Internal) repo() *Repo {
	return &Repo{}
}

func (_ *Internal) workflows() *Workflows {
	return &Workflows{}
}
