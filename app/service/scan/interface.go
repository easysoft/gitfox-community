package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

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

type IRule interface {
	GetByID(ctx context.Context, session *authModel.Session, id uint) (*scanModel.Rule, error)
	Migrate(ctx context.Context) error
	List(ctx context.Context, session *authModel.Session, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error)
	UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error
}
