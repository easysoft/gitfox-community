package git

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

type IGit interface {
	GitServicePack(ctx context.Context, session *auth.Session, repoRef string, options api.ServicePackOptions) error
}

type GitService struct {
}

func NewGitService() IGit {
	return &GitService{}
}

// GitServicePack executes the service pack part of git's smart http protocol (receive-/upload-pack).
func (s *GitService) GitServicePack(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	options api.ServicePackOptions,
) error {
	return nil
}
