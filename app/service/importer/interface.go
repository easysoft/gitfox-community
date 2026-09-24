// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package importer

import (
	"context"

	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
)

type Interface interface {
	Run(ctx context.Context, provider repoModel.Provider, repo *repoModel.Repo, cloneURL string, mirror bool, noticeLang string) error
	Create(ctx context.Context, session *auth.Session, req *repoModel.Repo) (*repoModel.Repo, error)
	LoadRepositoryFromProvider(ctx context.Context, provider repoModel.Provider, slug string) (RepositoryInfo, repoModel.Provider, error)
	GetProgress(ctx context.Context, repoID uint) (repoModel.Progress, error)
	CreateMirrorSyncJob(ctx context.Context, repoID int64) error
	GetMirrorSyncProgress(ctx context.Context, repoID int64) (repoModel.Progress, error)
}
