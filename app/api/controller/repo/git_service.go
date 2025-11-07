package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
)

// GitServicePack executes the service pack part of git's smart http protocol (receive-/upload-pack).
func (c *RepoController) GitServicePack(
	ctx context.Context,
	session *auth.Session,
	repoRef string,
	options api.ServicePackOptions,
) error {

	// todo
	return nil
}
