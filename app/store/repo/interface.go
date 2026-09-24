// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package repo

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/repo_store.go -package=mock

import (
	"context"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/database"
)

type RepoPageResult = *database.PageResult[*repoModel.Repo]

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error)
	Update(ctx context.Context, repo *Repository, id uint) (*repoModel.Repo, error)
	UpdateOptLock(ctx context.Context, repo *repoModel.Repo, mutateFn func(repo *repoModel.Repo) error) (*repoModel.Repo, error)
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
	Delete(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error
	GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error)
	List(ctx context.Context, filter repoModel.ReqFilter) (RepoPageResult, error)
	ListAllGitRepos(ctx context.Context) ([]*repoModel.Repo, error)
	GetMigrate(ctx context.Context) ([]*repoModel.Repo, error)
}
