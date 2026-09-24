// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package tag

import (
	"github.com/easysoft/gitfox-community/app/model/commit"
)

// ReqForCreate used for tag creation apis.
type ReqForCreate struct {
	Name    string `json:"name" binding:"required"`
	Source  string `json:"source"`
	Message string `json:"message"`
}

type ReqFilter struct {
	Query         string `json:"query" form:"query"`
	Sort          string `json:"sort" form:"sort" default:"name"`
	Order         string `json:"order" form:"order" default:"asc"`
	Page          int    `json:"page" form:"page" default:"1"`
	PageSize      int    `json:"pageSize" form:"pageSize" default:"20"`
	IncludeCommit bool   `json:"includeCommit" form:"includeCommit"`
}

type Tag struct {
	Name      string            `json:"name"`
	SHA       string            `json:"sha"`
	Message   string            `json:"message,omitempty"`
	CreatedBy *commit.Signature `json:"createdBy,omitempty"`
	Commit    *commit.Commit    `json:"commit,omitempty"`
}
