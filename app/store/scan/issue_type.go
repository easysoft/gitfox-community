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
	Title           string                        `gorm:"column:title"`
	Content         string                        `gorm:"column:content"`
	FilePath        string                        `gorm:"column:filePath"`
	OldFilePath     string                        `gorm:"column:oldFilePath"`
	StartLine       int64                         `gorm:"column:startLine"`
	EndLine         int64                         `gorm:"column:endLine"`
	OldCode         string                        `gorm:"column:oldCode"`
	NewCode         string                        `gorm:"column:newCode"`
	PpmID           int64                         `gorm:"column:ppmID"`
	RepoID          int64                         `gorm:"column:repoID"`
	RepoBranch      string                        `gorm:"column:repoBranch"`
	CreatedByTaskID int64                         `gorm:"column:createdByTaskID"`
	UpdatedByTaskID int64                         `gorm:"column:updatedByTaskID"`
	Status          constant.ScanIssueStatus      `gorm:"column:status"`
	ScanMethod      constant.ScanMethod           `gorm:"column:scanMethod"`
	Resolution      *constant.ScanIssueResolution `gorm:"column:resolution;default:''"`
	Payload         []byte                        `gorm:"column:payload"`
	CreatedDate     *datetime.Local               `gorm:"column:createdDate"`
	EditedDate      *datetime.Local               `gorm:"column:editedDate"`
	Deleted         bool                          `gorm:"column:deleted"`
	ResolvedDate    *datetime.Local               `gorm:"column:resolvedDate"`
	ClosedDate      *datetime.Local               `gorm:"column:closedDate"`
	IgnoredDate     *int64                        `gorm:"column:ignoredDate"`
	CreatedBy       string                        `gorm:"column:createdBy"`
	EditedBy        string                        `gorm:"column:editedBy"`
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
	Title           string                        `gorm:"column:title"`
	Content         string                        `gorm:"column:content"`
	FilePath        string                        `gorm:"column:filePath"`
	OldFilePath     string                        `gorm:"column:oldFilePath"`
	StartLine       int64                         `gorm:"column:startLine"`
	EndLine         int64                         `gorm:"column:endLine"`
	OldCode         string                        `gorm:"column:oldCode"`
	NewCode         string                        `gorm:"column:newCode"`
	PpmID           int64                         `gorm:"column:ppmID"`
	RepoID          int64                         `gorm:"column:repoID"`
	RepoBranch      string                        `gorm:"column:repoBranch"`
	Status          constant.ScanIssueStatus      `gorm:"column:status"`
	ScanMethod      constant.ScanMethod           `gorm:"column:scanMethod"`
	Resolution      *constant.ScanIssueResolution `gorm:"column:resolution"`
	Payload         []byte                        `gorm:"column:payload"`
	CreatedByTaskID int64                         `gorm:"column:createdByTaskID"`
	UpdatedByTaskID int64                         `gorm:"column:updatedByTaskID"`
	PlanID          int64                         `gorm:"column:planID"`
	PlanName        string                        `gorm:"column:planName"`
	CreatedDate     *datetime.Local               `gorm:"column:createdDate"`
	ResolvedDate    *datetime.Local               `gorm:"column:resolvedDate"`
	ClosedDate      *datetime.Local               `gorm:"column:closedDate"`
	IgnoredDate     *int64                        `gorm:"column:ignoredDate"`
	CreatedBy       string                        `gorm:"column:createdBy"`
	EditedBy        string                        `gorm:"column:editedBy"`
	TriggerID       int64                         `gorm:"column:triggerID"`
	TriggerType     constant.ScanPlanTriggerType  `gorm:"column:triggerType"`
}

func mapIssueJoinToModel(in *issueJoin) *scanModel.ScanIssueInfo {
	trigger := new(scanModel.ScanTaskTrigger)
	if in.TriggerType == constant.ScanPlanTriggerTypeAction || in.TriggerType == constant.ScanPlanTriggerTypeCron {
		trigger.ID = in.TriggerID
		trigger.Type = in.TriggerType
	} else {
		trigger.Type = constant.ScanPlanTriggerTypeManual
	}
	return &scanModel.ScanIssueInfo{
		ID:              in.ID,
		Key:             in.Key,
		RuleID:          in.RuleID,
		RuleName:        in.RuleName,
		RuleKey:         in.RuleKey,
		RulePriority:    in.RulePriority,
		RuleType:        in.RuleType,
		Title:           in.Title,
		Content:         in.Content,
		FilePath:        in.FilePath,
		OldFilePath:     in.OldFilePath,
		StartLine:       in.StartLine,
		EndLine:         in.EndLine,
		OldCode:         in.OldCode,
		NewCode:         in.NewCode,
		PpmID:           in.PpmID,
		RepoID:          in.RepoID,
		RepoBranch:      in.RepoBranch,
		Status:          in.Status,
		ScanMethod:      in.ScanMethod,
		Resolution:      in.Resolution,
		Payload:         json.RawMessage(in.Payload),
		CreatedByTaskID: in.CreatedByTaskID,
		UpdatedByTaskID: in.UpdatedByTaskID,
		PlanID:          in.PlanID,
		PlanName:        in.PlanName,
		CreatedDate:     in.CreatedDate,
		ResolvedDate:    in.ResolvedDate,
		ClosedDate:      in.ClosedDate,
		IgnoredDate:     in.IgnoredDate,
		CreatedBy:       in.CreatedBy,
		EditedBy:        in.EditedBy,
		Trigger:         trigger,
	}
}
