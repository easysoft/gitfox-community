package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/scan_store.go -package=mock

type IRule interface {
	ListByRuleSetID(ctx context.Context, ruleIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	ExcludeByRuleSetID(ctx context.Context, excludeIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	ListByIDs(ctx context.Context, ruleIDs []int64) ([]*scanModel.Rule, error)

	GetByID(ctx context.Context, id uint) (*scanModel.Rule, error)
	UpdateStatus(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error)
	AllScanBaseByField(ctx context.Context, field string) ([]scanModel.ScanBaseResp, error)
	Migrate(ctx context.Context, rulesFile *resources.RulesFile) error
}

type IRuleSet interface {
	Create(ctx context.Context, ruleSet *RuleSet) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) (*scanModel.RuleSet, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.RuleSetReqFilter) (*database.PageResult[*scanModel.RuleSet], error)
	GetRuleIDList(ctx context.Context, ruleSetID uint) ([]int64, error)
	IsUsedBySolution(ctx context.Context, id uint) (bool, error)
	BindRule(ctx context.Context, ruleSetID uint, ruleIDs []uint) error
	UnBindRule(ctx context.Context, ruleSetID uint, ruleIDs []uint) error
}

// ISolution 扫描方案 store 接口
type ISolution interface {
	Create(ctx context.Context, solution *SolutionRules) (*scanModel.Solution, error)
	GetByID(ctx context.Context, id uint) (*scanModel.Solution, error)
	Find(ctx context.Context, name string) (*scanModel.Solution, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.SolutionReqFilter) (*database.PageResult[*scanModel.Solution], error)
	GetRulesetIDList(ctx context.Context, solutionID uint) ([]int64, error)
	BindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error
	UnbindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error
	ListRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
	ListUnbindRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
	GetSolutionsByPlanID(ctx context.Context, planID uint) ([]*scanModel.Solution, []int64, error)
	IsUsedByPlan(ctx context.Context, solutionID uint) (bool, error)
}

// IScanPlan 扫描计划 store 接口
type IScanPlan interface {
	Create(ctx context.Context, plan *ScanPlan) (*scanModel.ScanPlan, error)
	GetByID(ctx context.Context, id uint) (*scanModel.ScanPlan, error)
	GetByRepoIDAndID(ctx context.Context, repoID uint, id uint) (*scanModel.ScanPlan, error)
	FindByName(ctx context.Context, repoID uint, name string) (*scanModel.ScanPlan, error)
	Update(ctx context.Context, plan *ScanPlan) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error)
	GetSolutionIDList(ctx context.Context, planID uint) ([]int64, error)
	BindSolutions(ctx context.Context, planID uint, solutionIDs []uint) error
	UnbindSolutions(ctx context.Context, planID uint, solutionIDs []uint) error
	ListUnbindSolutions(ctx context.Context, planID uint) ([]*scanModel.Solution, error)
}

// IScanTrigger 触发器 store 接口
type IScanTrigger interface {
	Create(ctx context.Context, trigger *ScanPlanTrigger) (*scanModel.ScanTrigger, error)
	GetByID(ctx context.Context, id uint) (*scanModel.ScanTrigger, error)
	GetByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error)
	Update(ctx context.Context, trigger *ScanPlanTrigger) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, repoID uint, planID uint, filter scanModel.ScanTriggerReqFilter) (*database.PageResult[*scanModel.ScanTrigger], error)
	Count(ctx context.Context, repoID uint, planID uint) (int64, error)
	FindByRepoIDAndActions(ctx context.Context, repoID uint, actions []constant.TriggerAction) ([]*scanModel.ScanTrigger, error)
}

// IScanTriggerSolution 触发器关联解决方案 store 接口
type IScanTriggerSolution interface {
	ReplaceBatch(ctx context.Context, triggerID uint64, solutionIDs []int64, createdBy string) error
	ListByTriggerID(ctx context.Context, triggerID uint64) ([]int64, error)
	ListByTriggerIDs(ctx context.Context, triggerIDs []uint64) (map[uint64][]int64, error)
	DeleteByTriggerID(ctx context.Context, triggerID uint64) error
}

// IScanCondition 触发器条件 store 接口
//
// Note: ops_scan_plan_conditions 同时承载 trigger-scoped 与 plan-scoped 条件。
// trigger-scoped 使用 IScanCondition；plan-scoped 使用 IScanPlanCondition.
type IScanCondition interface {
	BatchCreate(ctx context.Context, conditions []*ScanPlanCondition) error
	BatchDelete(ctx context.Context, repoID uint, planID uint, triggerID uint, ids []uint) error
	List(ctx context.Context, triggerID uint) ([]*scanModel.ScanCondition, error)
	ListByPlanID(ctx context.Context, planID uint) ([]*scanModel.ScanCondition, error)
}

// IScanPlanCondition 扫描计划级条件 store 接口（triggerID=0）
//
// This is intentionally separate from IScanCondition to avoid changing existing
// trigger-scoped condition APIs.
type IScanPlanCondition interface {
	Create(ctx context.Context, condition *ScanPlanCondition) (*scanModel.ScanCondition, error)
	GetByID(ctx context.Context, id uint) (*scanModel.ScanCondition, error)
	Update(ctx context.Context, condition *ScanPlanCondition) error
	Delete(ctx context.Context, id uint) error
	ListByPlan(ctx context.Context, repoID, planID uint) ([]*scanModel.ScanCondition, error)
	BatchCreateByPlan(ctx context.Context, conditions []*ScanPlanCondition) error
	BatchDeleteByIDs(ctx context.Context, repoID, planID uint, ids []uint) error
}

// IScanIssue 扫描问题 store 接口
type IScanIssue interface {
	Create(ctx context.Context, issue *ScanIssue) error
	BatchCreate(ctx context.Context, issues []*ScanIssue) error
	Get(ctx context.Context, id uint) (*scanModel.ScanIssueInfo, error)
	List(ctx context.Context, filter scanModel.ScanIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error)
	ListByIDs(ctx context.Context, ids []int64) ([]*scanModel.ScanIssueInfo, error)
	UpdateStatus(ctx context.Context, ids []int64, status constant.ScanIssueStatus) error
	UpdateResolution(ctx context.Context, ids []int64, resolution *constant.ScanIssueResolution) error
	UpdateTimestamps(ctx context.Context, in scanModel.ScanIssueUpdateStatusReq) error
	ResetExpiredIgnore(ctx context.Context, ids []int64) error
	ListKeysByTask(ctx context.Context, repoID int64, branch string, taskID uint, method constant.ScanMethod) ([]*scanModel.ScanIssueKey, error)
	UpdateByTaskID(ctx context.Context, taskID uint, issueIDs []int64) error
	DisappearedByTask(ctx context.Context, taskID uint, issueIDs []int64) error
	CreateIssueBindByTaskID(ctx context.Context, taskID uint, issueKeys []string) error
	ListByTaskForResult(ctx context.Context, taskID uint, planID int64) ([]*scanModel.Rule, error)
	ListByTaskID(ctx context.Context, taskID uint, filter scanModel.ScanTaskIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error)
	ListDistinctFilesByRepo(ctx context.Context, repoID int64, taskFilter *scanModel.IssueTaskFilter) ([]*scanModel.IssueFileRaw, error)
	ListDistinctRulesByRepo(ctx context.Context, repoID int64, taskFilter *scanModel.IssueTaskFilter) ([]*scanModel.IssueRuleRaw, error)
}

// IScanTask 扫描任务 store 接口
type IScanTask interface {
	Create(ctx context.Context, task *ScanTask) (*scanModel.ScanTask, error)
	GetByID(ctx context.Context, id uint) (*scanModel.ScanTask, error)
	FindByExecID(ctx context.Context, execID uint) (*scanModel.ScanTask, error)
	GetTaskInfo(ctx context.Context, taskID uint) (*scanModel.ScanTaskInfo, error)
	List(ctx context.Context, planID uint) ([]*scanModel.ScanTask, error)
	ListTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (*database.PageResult[*scanModel.ScanTaskInfo], error)
	ListBatch(ctx context.Context, repoID int64, batchSize int, fn func(task *scanModel.ScanTaskInfo) error) error
	CountTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (int64, error)
	Update(ctx context.Context, task *ScanTask) error
	UpdateResult(ctx context.Context, taskID uint, result constant.ScanTaskResult) error
	UpdateIssueNumber(ctx context.Context, taskID uint, issueNumber int64) error
	Delete(ctx context.Context, id uint) error
	NextPlanNumber(ctx context.Context, planID uint64) (int64, error)
	NextRepoNumber(ctx context.Context, repoID int64) (int64, error)
}
