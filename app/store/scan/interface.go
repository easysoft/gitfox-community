package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

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

type IRule interface {
	ListByRuleSetID(ctx context.Context, ruleIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)
	ExcludeByRuleSetID(ctx context.Context, excludeIDs []int64, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error)

	GetByID(ctx context.Context, id uint) (*scanModel.Rule, error)
	UpdateStatus(ctx context.Context, id uint) error
	List(ctx context.Context, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error)
	Migrate(ctx context.Context, rulesFile *resources.RulesFile) error
}
