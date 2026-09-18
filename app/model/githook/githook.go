package githook

import (
	"github.com/easysoft/gitfox-community/pkg/git/hook"
)

// GithookInputBase contains the base input of the githook apis.
type GithookInputBase struct {
	RepoID           int64
	PrincipalAccount string
	Internal         bool // Internal calls originate from Gitness, and external calls are direct git pushes.
}

// GithookPreReceiveInput is the input for the pre-receive githook api call.
type GithookPreReceiveInput struct {
	GithookInputBase
	hook.PreReceiveInput
}

// GithookUpdateInput is the input for the update githook api call.
type GithookUpdateInput struct {
	GithookInputBase
	hook.UpdateInput
}

// GithookPostReceiveInput is the input for the post-receive githook api call.
type GithookPostReceiveInput struct {
	GithookInputBase
	hook.PostReceiveInput
}
