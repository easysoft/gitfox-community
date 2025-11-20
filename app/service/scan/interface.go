package scan

import (
	"context"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
)

type IRuleSet interface {
	Create(ctx context.Context, req scanModel.RuleSetCreateReq) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, id uint, req scanModel.RuleSetUpdateReq) (*scanModel.RuleSet, error)
}
