package version

import (
	"fmt"
	"runtime"
)

var (
	major      = "0"
	minor      = "0"
	gitVersion = "v0.0.0-dev"
	gitCommit  = ""
	commitDate = "1970-01-01T00:00:00Z"
)

// Version is a struct for version information.
type Version struct {
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	CommitDate string `json:"commitDate"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

// GetVersion returns this binary's version.
func GetVersion() Version {
	return Version{
		Major:      major,
		Minor:      minor,
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		CommitDate: commitDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
