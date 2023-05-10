package builder

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for a container.
type Config struct {
	base `yaml:",inline"`

	// Label is the label of the container. Optional. Default is "ubuntu-latest".
	Label string `yaml:"label,omitempty"`

	// Modifications is a list of modifications to be applied to the container. Optional. Default is no modifications.
	Modifications []Modification `yaml:"modifications,omitempty"`
}

// base holds the base configuration for a container. It is used by the builder to create a container. It should not
// contain more than one of the following fields: From, DockerFile, ImportPath. If more than one of these fields are
// present, the builder will override the previous value.
type base struct {
	// Creates a container from a Docker image.
	From string `yaml:"from,omitempty"`

	// Context is path to the context directory for the Dockerfile.
	Context string `yaml:"context,omitempty"`

	// Builds a container from a Dockerfile.
	DockerFile string `yaml:"dockerfile,omitempty"`

	// Imports a container from a tar file.
	ImportPath string `yaml:"import,omitempty"`
}

// Modification represents a single modification to be applied to a container.
type Modification struct {
	// Name of the modification. Optional.
	Name string `yaml:"name,omitempty"`

	// Run script to be executed in the step.
	Run string `yaml:"run,omitempty"`

	// Apt packages to be installed in the step. All packages will be installed in a single step to reduce the number
	// of layers in the resulting image.
	AptPackages AptPackages `yaml:"apt-packages,omitempty"`

	// Tools to be installed in the step.
	Tools []string `yaml:"tools,omitempty"`
}

// AptPackages is a list of apt packages to be installed. It can be a list of strings or a list of objects. This
// requires custom unmarshalling to support both formats.
var _ yaml.Unmarshaler = new(AptPackages)

// AptPackages represents a list of apt packages to be installed.
type AptPackages []AptPackage

// AptPackage represents an apt package to be installed.
type AptPackage struct {
	// Name is the name of the apt package.
	Name string

	// Version is the version of the apt package. Optional.
	Version string
}

// String returns the string representation of the apt package.
func (a *AptPackage) String() string {
	if a.Version != "" {
		return a.Name + "=" + a.Version
	}
	return a.Name
}

// UnmarshalYAML unmarshal the YAML value into an AptPackages. It supports both string and object formats.
//
// Example:
//
//	packages: [git, curl]
//	packages:
//	  - name: git
//	    version: 2.30.1-1ubuntu1.2
//	  - name: curl
//	    version: 7.68.0-1ubuntu2.7
func (a *AptPackages) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("apt packages must be a sequence")
	}

	for _, item := range value.Content {
		if item.Kind == yaml.ScalarNode {
			*a = append(*a, AptPackage{Name: item.Value})
			continue
		}

		if item.Kind != yaml.MappingNode {
			var apt AptPackage

			if err := value.Decode(&apt); err != nil {
				return err
			}

			*a = append(*a, apt)
		}

		return fmt.Errorf("apt package must be a string or an AptPackage object")
	}

	return nil
}
