CMD  := $(CURDIR)/hack/mage
ARGS ?= run

default: mage

mage:
	@$(CMD) $(ARGS)

help:
	@echo "This Makefile is a wrapper for mage. It is intended to transition from make to mage easily."
	@echo "Usage: make [target] [ARGS=...]"
	@echo "Targets:"
	@echo "  mage:    Runs mage with the default arguments. This is the default target."
	@echo "  help:    Prints this help message."
	@echo ""
	@echo "Arguments:"
	@echo "  ARGS:    Arguments to pass to mage. Defaults to \"run\"."