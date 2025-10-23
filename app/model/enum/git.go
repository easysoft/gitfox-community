package enum

import (
	"fmt"
	"strings"
)

// GitServiceType represents the different types of service values send by git's smart http protocol.
// See https://git-scm.com/docs/http-protocol#_smart_clients for more details.
type GitServiceType string

const (
	// GitServiceTypeReceivePack is sent by git push operations (server "receives" data from client).
	GitServiceTypeReceivePack GitServiceType = "receive-pack"
	// GitServiceTypeUploadPack is sent by git pull operations (server "uploads" data to client).
	GitServiceTypeUploadPack GitServiceType = "upload-pack"
)

// ParseGitServiceType parses the git service type string and returns the equivalent enumeration.
// If the value is unknown and doesn't represent a git service type, an error is returned.
func ParseGitServiceType(s string) (GitServiceType, error) {
	switch strings.ToLower(s) {
	case string(GitServiceTypeReceivePack):
		return GitServiceTypeReceivePack, nil
	case string(GitServiceTypeUploadPack):
		return GitServiceTypeUploadPack, nil
	default:
		return "", fmt.Errorf("unknown git service type provided: %q", s)
	}
}
