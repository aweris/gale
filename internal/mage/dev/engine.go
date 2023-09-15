package dev

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Engine mg.Namespace

const (
	RegistryContainerName           = "registry"
	RegistryContainerPort           = "5000"
	DaggerEngineContainerNamePrefix = "dagger-engine"
	DaggerEngineVersion             = "v0.8.5"
	EngineToml                      = `debug = true

insecure-entitlements = ["security.insecure"]

[registry."docker.io"]
    mirrors = ["{{.Registry}}"]

[registry."ghcr.io"]
	mirrors = ["{{.Registry}}"]

[registry."{{.Registry}}"]
    insecure = true
    http = true
`
)

// Start starts the dagger engine and the registry container with mirroring enabled.
func (_ Engine) Start() error {
	mg.Deps(Engine.Clean)

	err := sh.Run("docker", "run", "-d", "--name", RegistryContainerName, "-p", fmt.Sprintf("%s:%s", RegistryContainerPort, RegistryContainerPort), "registry:2")
	if err != nil {
		return err
	}

	ip, err := sh.Output("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", RegistryContainerName)
	if err != nil {
		return err
	}

	config := strings.ReplaceAll(EngineToml, "{{.Registry}}", fmt.Sprintf("%s:%s", ip, RegistryContainerPort))

	err = os.WriteFile(filepath.Join(os.Getenv("PWD"), "engine.toml"), []byte(config), 0600)
	if err != nil {
		return err
	}

	return sh.Run("docker", "run", "-d", "--name", EngineContainerName(), "-v", fmt.Sprintf("%s:/etc/dagger/engine.toml", filepath.Join(os.Getenv("PWD"), "engine.toml")), "--privileged", fmt.Sprintf("registry.dagger.io/engine:%s", DaggerEngineVersion))
}

// Env prints the environment variables for running gale with development dagger engine.
func (_ Engine) Env() error {
	fmt.Printf("export _EXPERIMENTAL_DAGGER_RUNNER_HOST=docker-container://%s\n", EngineContainerName())

	ip, err := sh.Output("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", RegistryContainerName)
	if err != nil {
		return err
	}

	fmt.Printf("export _GALE_DOCKER_REGISTRY=%s:%s\n", ip, RegistryContainerPort)

	return nil
}

// Clean force removes registry and dagger engine containers
func (_ Engine) Clean() error {

	sh.Run("docker", "rm", "-f", RegistryContainerName)
	sh.Run("docker", "rm", "-f", EngineContainerName())

	return nil
}

func EngineContainerName() string {
	return strings.Join([]string{DaggerEngineContainerNamePrefix, DaggerEngineVersion}, "-")
}
