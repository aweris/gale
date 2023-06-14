WORKFLOW ?= example-golangci-lint
JOB      ?= golangci-lint

ARGS ?= run $(WORKFLOW) $(JOB) --export --disable-checkout

DAGGER_CMD     := ./bin/dagger
DAGGER_VERSION := v0.6.1

default: run

run: $(DAGGER_CMD);
	$(shell $(DAGGER_CMD) run go run main.go $(ARGS))

help:
	@echo "This Makefile is a wrapper for dagger run to make it easier to run gale."
	@echo "Usage: make [target] [ARGS=...]"
	@echo "Targets:"
	@echo "  run:     Runs main.go with the given arguments."
	@echo "  help:    Prints this help message."
	@echo ""
	@echo "Arguments:"
	@echo "  ARGS:     Arguments to pass to gale. Defaults to \"$(ARGS)\"."
	@echo "  WORKFLOW: Path to the workflow file to use. Defaults to \"$(WORKFLOW)\"."
	@echo "  JOB:      Name of the job to run. Defaults to \"$(JOB)\"."

$(DAGGER_CMD):
	$(shell curl -L https://dl.dagger.io/dagger/install.sh | VERSION=$(DAGGER_VERSION) bash)
