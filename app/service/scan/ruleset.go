package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
)

var _ IRuleSet = (*RuleSetService)(nil)

type RuleSetService struct {
	ruleSetStore scanStore.IRuleSet
}

func NewRuleSetService(ruleSetStore scanStore.IRuleSet) *RuleSetService {
	return &RuleSetService{
		ruleSetStore: ruleSetStore,
	}
}

func (s *RuleSetService) Create(ctx context.Context, session *authModel.Session, req scanModel.RuleSetCreateReq) (*scanModel.RuleSet, error) {
	// 将请求转换为 store 层的 RuleSet 结构
	ruleSet := &scanStore.RuleSet{
		Name:      req.Name,
		Desc:      req.Desc,
		DescrEn:   req.DescEn,
		Lang:      constant.ScanLangName(req.Lang),
		Tag:       req.Tag,
		Type:      constant.ScanRuleType(req.Type),
		Plugin:    req.Plugin,
		Status:    constant.ScanRuleStatusEnabled,
		CreatedBy: session.User.UID,
		EditedBy:  session.User.UID,
	}

	return s.ruleSetStore.Create(ctx, ruleSet)
}

func (s *RuleSetService) GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error) {
	return s.ruleSetStore.GetByID(ctx, id)
}

func (s *RuleSetService) Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.RuleSetUpdateReq) (*scanModel.RuleSet, error) {
	// 将请求转换为 store 层的 RuleSet 结构
	ruleSet := &scanStore.RuleSet{EditedBy: session.User.UID}

	if req.Name != "" {
		ruleSet.Name = req.Name
	}
	if req.Desc != "" {
		ruleSet.Desc = req.Desc
	}
	if req.DescEn != "" {
		ruleSet.DescrEn = req.DescEn
	}
	if req.Lang != "" {
		ruleSet.Lang = constant.ScanLangName(req.Lang)
	}
	if req.Tag != "" {
		ruleSet.Tag = req.Tag
	}
	if req.Type != "" {
		ruleSet.Type = constant.ScanRuleType(req.Type)
	}
	if req.Plugin != "" {
		ruleSet.Plugin = req.Plugin
	}
	if req.Status != "" {
		ruleSet.Status = constant.ScanRuleStatus(req.Status)
	}

	return s.ruleSetStore.Update(ctx, ruleSet, id)
}
