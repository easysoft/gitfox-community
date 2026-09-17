//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/statistics_store.go -package=mock

package statistics

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	// ListGitRepos 获取所有未删除的git代码库。
	ListGitRepos(ctx context.Context) ([]*repoModel.Repo, error)
}
