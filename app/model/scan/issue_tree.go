// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/constant"
)

// TreeNode 树节点
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Children []*TreeNode `json:"children,omitempty"`
	IsLeaf   bool        `json:"isLeaf"`
	Ref      any         `json:"ref,omitempty"`
	Count    int         `json:"issueCount,omitempty"`
}

// IssueTreeFilter issue 树查询过滤条件
type IssueTreeFilter struct {
	RepoID      int64                 `json:"repoID" form:"repoID"`
	TaskID      int64                 `json:"taskID" form:"taskID"`
	ScanMethods []constant.ScanMethod `json:"scanMethods,omitempty" form:"scanMethods"`
}

// IssueTaskFilter 任务维度过滤（用于 store 层查询）
type IssueTaskFilter struct {
	TaskID int64
	PlanID int64
	RepoID int64
	Branch string
}

// IssueFileRaw 文件维度聚合原始数据
type IssueFileRaw struct {
	Branch string `gorm:"column:repoBranch"`
	File   string `gorm:"column:path"`
	Count  int    `gorm:"column:totalCount"`
}

// IssueRuleRaw 规则维度聚合原始数据
type IssueRuleRaw struct {
	ID     int64                 `gorm:"column:id"`
	Name   string                `gorm:"column:name"`
	Lang   constant.ScanLangName `gorm:"column:lang"`
	Plugin string                `gorm:"column:plugin"`
	Type   constant.ScanRuleType `gorm:"column:type"`
	Count  int                   `gorm:"column:totalCount"`
}
