// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
