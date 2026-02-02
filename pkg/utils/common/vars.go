package common

import (
	"fmt"
)

var (
	Version       string
	BuildDate     string
	GitCommitHash string
	GitBranch     string
)

// GetVersionWithHash returns the version string with the git commit hash and build date.
func GetVersionWithHash() string {
	return fmt.Sprintf("%s-%s-%s", Version, GitCommitHash, BuildDate)
}
