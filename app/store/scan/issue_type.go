package scan

import (
	"encoding/json"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// ScanIssue 扫描问题 DB 模型
type ScanIssue struct {
	ID              uint64                        `gorm:"column:id;primaryKey"`
	Key             string                        `gorm:"column:issueKey"`
	RuleID          int64                         `gorm:"column:ruleID"`
	Message         string                        `gorm:"column:message"`
	Path            string                        `gorm:"column:path"`
	Line            int64                         `gorm:"column:line"`
	RepoID          int64                         `gorm:"column:repoID"`
	RepoBranch      string                        `gorm:"column:repoBranch"`
	CreatedByTaskID int64                         `gorm:"column:createdByTaskID"`
	UpdatedByTaskID int64                         `gorm:"column:updatedByTaskID"`
	Status          constant.ScanIssueStatus      `gorm:"column:status"`
	ScanMethod      constant.ScanMethod           `gorm:"column:scanMethod"`
	Resolution      *constant.ScanIssueResolution `gorm:"column:resolution"`
	Payload         []byte                        `gorm:"column:payload"`
	CreatedDate     *datetime.Local               `gorm:"column:createdDate"`
	EditedDate      *datetime.Local               `gorm:"column:editedDate"`
	Deleted         bool                          `gorm:"column:deleted"`
}

func (r *ScanIssue) TableName() string { return "ops_scan_issues" }

func (r *ScanIssue) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

func (r *ScanIssue) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

// ScanIssueTaskBind 问题与任务绑定关系 DB 模型
type ScanIssueTaskBind struct {
	ID      uint64 `gorm:"column:id;primaryKey"`
	TaskID  int64  `gorm:"column:taskID"`
	IssueID int64  `gorm:"column:issueID"`
}

func (r *ScanIssueTaskBind) TableName() string { return "ops_scan_issue_task_binds" }

// issueJoin issue 联表查询中间结构
type issueJoin struct {
	ID              int64                         `gorm:"column:id"`
	Key             string                        `gorm:"column:issueKey"`
	RuleID          int64                         `gorm:"column:ruleID"`
	RuleName        string                        `gorm:"column:ruleName"`
	RuleKey         string                        `gorm:"column:ruleKey"`
	RulePriority    constant.ScanRulePriority     `gorm:"column:rulePriority"`
	RuleType        constant.ScanRuleType         `gorm:"column:ruleType"`
	Message         string                        `gorm:"column:message"`
	Path            string                        `gorm:"column:path"`
	Line            int64                         `gorm:"column:line"`
	RepoID          int64                         `gorm:"column:repoID"`
	RepoBranch      string                        `gorm:"column:repoBranch"`
	Status          constant.ScanIssueStatus      `gorm:"column:status"`
	ScanMethod      constant.ScanMethod           `gorm:"column:scanMethod"`
	Resolution      *constant.ScanIssueResolution `gorm:"column:resolution"`
	Payload         []byte                        `gorm:"column:payload"`
	CreatedByTaskID int64                         `gorm:"column:createdByTaskID"`
	UpdatedByTaskID int64                         `gorm:"column:updatedByTaskID"`
	CreatedDate     *datetime.Local               `gorm:"column:createdDate"`
}

func mapIssueJoinToModel(in *issueJoin) *scanModel.ScanIssueInfo {
	return &scanModel.ScanIssueInfo{
		ID:              in.ID,
		Key:             in.Key,
		RuleID:          in.RuleID,
		RuleName:        in.RuleName,
		RuleKey:         in.RuleKey,
		RulePriority:    in.RulePriority,
		RuleType:        in.RuleType,
		Message:         in.Message,
		Path:            in.Path,
		Line:            in.Line,
		RepoID:          in.RepoID,
		RepoBranch:      in.RepoBranch,
		Status:          in.Status,
		ScanMethod:      in.ScanMethod,
		Resolution:      in.Resolution,
		Payload:         json.RawMessage(in.Payload),
		CreatedByTaskID: in.CreatedByTaskID,
		UpdatedByTaskID: in.UpdatedByTaskID,
		CreatedDate:     in.CreatedDate,
	}
}
