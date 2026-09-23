// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pullrequest

import (
	"context"
	"net/http"

	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/model/commit"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/git"
)

type Interface interface {
	GetMetadata(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint) (*repoModel.DiffStats, error)
	GetDiffStats(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.DiffStats, error)
	MergeCheck(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.MergeCheck, error)

	RawDiff(ctx context.Context, w http.ResponseWriter, session *auth.Session, repoID uint, pullreqID uint, setSHAs func(sourceSHA, mergeBaseSHA string), files ...repoModel.FileDiff) error
	Diff(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, setSHAs func(sourceSHA, mergeBaseSHA string), includePatch bool, files ...repoModel.FileDiff) (resp.Stream[*git.FileDiff], error)
	Commits(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, page, pageSize uint) ([]commit.Commit, int, error)
	Merge(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, params repoModel.Merge) (*repoModel.MergeResponse, error)
	TriggerEvent(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, params repoModel.EventTrigger) error
}
