package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authz"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
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

func (s *RuleService) Migrate(ctx context.Context) error {
	// 加载所有规则文件
	rulesFiles, err := resources.LoadAllRulesFiles()
	if err != nil {
		return errors.Wrapf(err, "failed to load rules files")
	}

	// 逐个文件迁移
	for _, file := range rulesFiles {
		if err := s.ruleStore.Migrate(ctx, file); err != nil {
			return errors.Wrapf(err, "failed to migrate rules for plugin %s", file.Plugin)
		}
	}

	return nil
}

func (s *RuleService) UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error {
	return s.ruleStore.UpdateStatus(ctx, id)
}

func (s *RuleService) List(ctx context.Context, session *authModel.Session, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error) {
	return s.ruleStore.List(ctx, filter)
}
