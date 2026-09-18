package pullrequest

import (
	"context"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
)

// PushRequestCleaner removes temporary branches and references created by the
// paid "code push request" flow once a pull request is merged or closed.
//
// The community edition binds NoopPushRequestCleaner; the paid edition provides
// the implementation that cleans up ztflow temporary branches and AGit
// refs/for/* references.
type PushRequestCleaner interface {
	// Cleanup removes temporary branches/refs for the given PR. It returns
	// whether the source branch was deleted. Implementations decide whether the
	// PR belongs to the push request flow.
	Cleanup(ctx context.Context, pr *repoModel.PullRequest) bool
}

// NoopPushRequestCleaner is the community-edition implementation of PushRequestCleaner.
type NoopPushRequestCleaner struct{}

var _ PushRequestCleaner = (*NoopPushRequestCleaner)(nil)

// NewNoopPushRequestCleaner creates the community-edition PushRequestCleaner.
func NewNoopPushRequestCleaner() *NoopPushRequestCleaner {
	return &NoopPushRequestCleaner{}
}

// Cleanup is a no-op in the community edition.
func (NoopPushRequestCleaner) Cleanup(_ context.Context, _ *repoModel.PullRequest) bool {
	return false
}
