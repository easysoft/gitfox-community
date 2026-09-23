// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package tag

import (
	"context"

	"github.com/easysoft/gitfox-community/app/model/auth"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	tagModel "github.com/easysoft/gitfox-community/app/model/tag"
)

type Interface interface {
	Create(ctx context.Context, req tagModel.ReqForCreate, session *auth.Session, repo *repoModel.Repo) (*tagModel.Tag, error)
	List(ctx context.Context, req tagModel.ReqFilter, repo repoModel.Repo) ([]tagModel.Tag, bool, error)
	Delete(ctx context.Context, name string, session *auth.Session, repo *repoModel.Repo) error
}
