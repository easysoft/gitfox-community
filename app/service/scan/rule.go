package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authz"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

var _ IRule = (*RuleService)(nil)

type RuleService struct {
	ruleStore  scanStore.IRule
	authorizer authz.Authorizer
}

func NewRuleService(ruleStore scanStore.IRule, authorizer authz.Authorizer) *RuleService {
	return &RuleService{
		ruleStore:  ruleStore,
		authorizer: authorizer,
	}
}

func (s *RuleService) GetByID(ctx context.Context, session *authModel.Session, id uint) (*scanModel.Rule, error) {
	return s.ruleStore.GetByID(ctx, id)
}

func (s *RuleService) Migrate(ctx context.Context, session *authModel.Session, files []string) error {
	return nil
}

func (s *RuleService) UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error {
	return s.ruleStore.UpdateStatus(ctx, id)
}

func (s *RuleService) List(ctx context.Context, session *authModel.Session, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error) {
	return s.ruleStore.List(ctx, filter)
}
