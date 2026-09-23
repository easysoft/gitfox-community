// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/auth/authz"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/resources"
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

func (s *RuleService) AllScanBaseByField(ctx context.Context, session *authModel.Session, field string) ([]scanModel.ScanBaseResp, error) {
	res, err := s.ruleStore.AllScanBaseByField(ctx, field)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].ID = int64(constant.ParseLang(res[i].Lang).ID())
		if res[i].Lang == "" || res[i].Lang == "Custom" {
			res[i].Lang = "Custom"
		}
	}

	return res, nil
}
