package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
)

type IRuleSet interface {
	Create(ctx context.Context, session *authModel.Session, req scanModel.RuleSetCreateReq) (*scanModel.RuleSet, error)
	GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error)
	Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.RuleSetUpdateReq) (*scanModel.RuleSet, error)
}
