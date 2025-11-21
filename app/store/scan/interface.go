package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type IRuleSet interface {
	Create(ctx context.Context, ruleSet *RuleSet) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, ruleSet *RuleSet, id uint) (*scanModel.RuleSet, error)
	Delete(ctx context.Context, id uint) error
	IsUsedBySolution(ctx context.Context, id uint) (bool, error)
	List(ctx context.Context, filter scanModel.RuleSetListReq) (*database.PageResult[*scanModel.RuleSet], error)
}

type IRule interface {
	ListByRuleSetID(ctx context.Context, ruleIDs []int64) ([]Rule, error)
}
