package main

import (
	"github.com/saschagrunert/demo"
)

func main() {
	d := demo.New()

	d.Setup(env.Setup)

	d.Add(list(), "list", "List all workflows and jobs under it for current repositories `main` branch")
	d.Add(run(), "run", "Run golangci-lint job from ci/workflows/lint workflow for aweris/gale repository default branch")
	d.Add(lintGoreleaser(), "lint-goreleaser", "Run golangci job from golangci-lint workflow for goreleaser/goreleaser repository tag v1.19.2")
	d.Add(testDagger(), "test-dagger", "Run sdk-go job from test workflow for dagger/dagger repository tag v0.8.1")

	d.Cleanup(env.Cleanup)

	d.Run()
}

// list returns a demo run for the list command for current repository
func list() *demo.Run {
	r := demo.NewRun("List Workflows")

	r.Step([]string{"List all available workflows under ci/workflows directory"}, env.RunGaleWithDagger("list --branch main"))

	return r
}

func run() *demo.Run {
	r := demo.NewRun("Run Workflow")

	r.Step([]string{"Contents of the workflow file"}, demo.S("curl https://raw.githubusercontent.com/aweris/gale/main/ci/workflows/lint.yaml"))

	r.Step([]string{"Run the workflow from custom directory for current repository"}, env.RunGaleWithDagger("run --repo aweris/gale --workflows-dir ci/workflows ci/workflows/lint.yaml golangci-lint"))

	return r
}

func lintGoreleaser() *demo.Run {
	r := demo.NewRun("Run golangci job from golangci-lint workflow for goreleaser v1.19.2")

	r.Step([]string{"Contents of the workflow file"}, demo.S("curl https://raw.githubusercontent.com/goreleaser/goreleaser/v1.19.2/.github/workflows/lint.yml"))

	r.Step([]string{"Run the workflow from custom directory for goreleaser/goreleaser repository"}, env.RunGaleWithDagger("run --repo goreleaser/goreleaser --tag v1.19.2 golangci-lint golangci"))

	return r
}

func testDagger() *demo.Run {
	r := demo.NewRun("Run sdk-go job from test workflow for dagger v0.8.1")

	r.Step([]string{"Contents of the workflow file"}, demo.S("curl https://raw.githubusercontent.com/dagger/dagger/v0.8.1/.github/workflows/test.yml"))

	r.Step([]string{"Run the workflow from custom directory for dagger/dagger repository"}, env.RunGaleWithDagger("run --repo dagger/dagger --tag v0.8.1 test sdk-go"))

	return r
}
