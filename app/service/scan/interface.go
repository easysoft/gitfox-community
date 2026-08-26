package scan

import (
	"context"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// CronScheduler is the interface used by the scan service to manage
// recurring cron jobs in the job scheduler. It is defined locally to avoid
// an import cycle with the jobs package.
type CronScheduler interface {
	AddRecurringWithData(ctx context.Context, jobName, jobType, jobData, cronDef string, maxDur time.Duration) error
	RemoveRecurring(ctx context.Context, jobName string) error
}

type IRule interface {
	GetByID(ctx context.Context, session *authModel.Session, id uint) (*scanModel.Rule, error)
	Migrate(ctx context.Context) error
	List(ctx context.Context, session *authModel.Session, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error)
	AllScanBaseByField(ctx context.Context, session *authModel.Session, field string) ([]scanModel.ScanBaseResp, error)
	UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error
}

type IRuleSet interface {
	Create(ctx context.Context, session *authModel.Session, req scanModel.RuleSetReqForCreate) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.RuleSetReqForUpdate) (*scanModel.RuleSet, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.RuleSetReqFilter) (*database.PageResult[*scanModel.RuleSet], error)
	GetBindRule(ctx context.Context, ruleSetID uint, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	GetUnBindRule(ctx context.Context, ruleSetID uint, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	BindRule(ctx context.Context, ruleSetID uint, req scanModel.RuleSetReqForBind) error
	UnBindRule(ctx context.Context, ruleSetID uint, req scanModel.RuleSetReqForBind) error
}

// ISolution 扫描方案 service 接口
type ISolution interface {
	Create(ctx context.Context, session *authModel.Session, req scanModel.SolutionReqForCreate) (*scanModel.Solution, error)
	GetByID(ctx context.Context, id uint) (*scanModel.Solution, error)
	Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.SolutionReqForUpdate) (*scanModel.Solution, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.SolutionReqFilter) (*database.PageResult[*scanModel.Solution], error)
	UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error
	ListRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
	ListUnbindRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
	BindRulesets(ctx context.Context, solutionID uint, req scanModel.SolutionReqForBind) error
	UnbindRulesets(ctx context.Context, solutionID uint, req scanModel.SolutionReqForBind) error
}

// IScanPlan 扫描计划 service 接口
type IScanPlan interface {
	Create(ctx context.Context, session *authModel.Session, repoID uint, req scanModel.ScanPlanReqForCreate) (*scanModel.ScanPlan, error)
	GetByID(ctx context.Context, repoID uint, id uint) (*scanModel.ScanPlan, error)
	Update(ctx context.Context, session *authModel.Session, repoID uint, id uint, req scanModel.ScanPlanReqForUpdate) (*scanModel.ScanPlan, error)
	Delete(ctx context.Context, repoID uint, id uint) error
	List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error)
	ListRules(ctx context.Context, repoID uint, planID uint, triggerID uint) (*scanModel.ScanPlan, []*scanModel.Rule, error)
	BindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error
	UnbindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error
	ListUnbindSolutions(ctx context.Context, repoID uint, planID uint) ([]*scanModel.Solution, error)
}

// IScanTrigger 触发器 service 接口
type IScanTrigger interface {
	Create(ctx context.Context, session *authModel.Session, planID uint, repoID uint, req scanModel.ScanTriggerReqForCreate) (*scanModel.ScanTrigger, error)
	GetByID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error)
	Update(ctx context.Context, session *authModel.Session, repoID uint, planID uint, id uint, req scanModel.ScanTriggerReqForUpdate) (*scanModel.ScanTrigger, error)
	Delete(ctx context.Context, repoID uint, planID uint, id uint) error
	List(ctx context.Context, repoID uint, planID uint, filter scanModel.ScanTriggerReqFilter) (*database.PageResult[*scanModel.ScanTrigger], error)
	FindByRepoIDAndActions(ctx context.Context, repoID uint, actions []constant.TriggerAction) ([]*scanModel.ScanTrigger, error)
}

// IScanCondition 触发器条件 service 接口
// NOTE: this is trigger-scoped conditions.
type IScanCondition interface {
	BatchCreate(ctx context.Context, session *authModel.Session, planID uint, triggerID uint, repoID uint, req scanModel.ScanConditionReqForBatchCreate) error
	BatchDelete(ctx context.Context, repoID uint, planID uint, triggerID uint, req scanModel.ScanConditionReqForBatchDelete) error
	List(ctx context.Context, repoID uint, planID uint, triggerID uint) ([]*scanModel.ScanCondition, error)
}

// IScanPlanCondition 计划级条件 service 接口（triggerID=0）
type IScanPlanCondition interface {
	Create(ctx context.Context, session *authModel.Session, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForCreate) (*scanModel.ScanCondition, error)
	List(ctx context.Context, repoID uint, planID uint) ([]*scanModel.ScanCondition, error)
	Find(ctx context.Context, repoID uint, planID uint, conditionID uint) (*scanModel.ScanCondition, error)
	Update(ctx context.Context, session *authModel.Session, repoID uint, planID uint, conditionID uint, req scanModel.ScanPlanConditionReqForUpdate) (*scanModel.ScanCondition, error)
	Delete(ctx context.Context, repoID uint, planID uint, conditionID uint) error
	BatchCreate(ctx context.Context, session *authModel.Session, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForBatchCreate) error
	BatchDelete(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForBatchDelete) error
}

// IScanTask 扫描任务 service 接口
type IScanTask interface {
	Create(ctx context.Context, in *ScanTaskInput) (*scanModel.ScanTask, error)
	CreateBranchHook(ctx context.Context, repo *repoModel.Repo, branch string, sha, beforeSha string, principal *authModel.Session) (*triggerer.Hook, error)
	Retry(ctx context.Context, session *authModel.Session, taskID uint, repo *repoModel.Repo) (*scanModel.ScanTask, error)
	GetTaskInfo(ctx context.Context, taskID uint) (*scanModel.ScanTaskInfo, error)
	ListTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (*database.PageResult[*scanModel.ScanTaskInfo], error)
	ListTaskIssues(ctx context.Context, taskID uint, filter scanModel.ScanTaskIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error)
	CountTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (int64, error)
	Delete(ctx context.Context, id uint) error
	SetupCron(ctx context.Context, trigger *scanModel.ScanTrigger) error
	RecycleCron(ctx context.Context, trigger *scanModel.ScanTrigger) error
	AnalyzeTaskUploadReport(ctx context.Context, repoID uint, taskID uint, method string, content []byte) (*scanModel.ScanIssueReport, error)
	GetIssueInfo(ctx context.Context, issueID uint) (*scanModel.ScanIssueInfo, error)
	ListIssueInfos(ctx context.Context, filter scanModel.ScanIssueReqFilter) (*database.PageResult[*scanModel.ScanIssueInfo], error)
	ListIssueInfosByIDs(ctx context.Context, ids []int64) ([]*scanModel.ScanIssueInfo, error)
	UpdateIssuesStatus(ctx context.Context, ids []int64, status constant.ScanIssueStatus) error
	SetIssuesStatus(ctx context.Context, in scanModel.ScanIssueUpdateStatusReq) error
	GetIssueFileTree(ctx context.Context, filter scanModel.IssueTreeFilter) (*scanModel.TreeNode, error)
	GetIssueRuleTree(ctx context.Context, repoID, taskID int64) (*scanModel.TreeNode, error)
}

type IAIReview interface {
	GetAIReviewRules(ctx context.Context) ([]*scanModel.AIReviewRule, error)
	TriggerCodeReview(ctx context.Context, in *CodeReviewInput) (*executionModel.Execution, error)
}
