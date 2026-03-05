package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

//go:generate mockgen -source=$GOFILE -destination=mock/_store.go -package=mock

type IRuleSet interface {
	Create(ctx context.Context, ruleSet *RuleSet) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, ruleSet *RuleSet, id uint) (*scanModel.RuleSet, error)
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
	Update(ctx context.Context, solution *SolutionRules) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.SolutionReqFilter) (*database.PageResult[*scanModel.Solution], error)
	GetRulesetIDList(ctx context.Context, solutionID uint) ([]int64, error)
	BindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error
	UnbindRulesets(ctx context.Context, solutionID uint, rulesetIDs []uint) error
	ListRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
	ListUnbindRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error)
}
type IRule interface {
	ListByRuleSetID(ctx context.Context, ruleIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	ExcludeByRuleSetID(ctx context.Context, excludeIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)

	GetByID(ctx context.Context, id uint) (*scanModel.Rule, error)
	UpdateStatus(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error)
	AllScanBaseByField(ctx context.Context, field string) ([]scanModel.ScanBaseResp, error)
	Migrate(ctx context.Context, rulesFile *resources.RulesFile) error
}
