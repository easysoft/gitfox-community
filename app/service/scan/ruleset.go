package scan

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

var _ IRuleSet = (*RuleSetService)(nil)

type RuleSetService struct {
	ruleSetStore scanStore.IRuleSet
	ruleStore    scanStore.IRule
}

func NewRuleSetService(ruleSetStore scanStore.IRuleSet, ruleStore scanStore.IRule) *RuleSetService {
	return &RuleSetService{
		ruleSetStore: ruleSetStore,
		ruleStore:    ruleStore,
	}
}

func (s *RuleSetService) Create(ctx context.Context, session *authModel.Session, req scanModel.RuleSetReqForCreate) (*scanModel.RuleSet, error) {
	// 将请求转换为 store 层的 RuleSet 结构
	ruleSet := &scanStore.RuleSet{
		Name:      req.Name,
		Desc:      req.Desc,
		DescEn:    req.DescEn,
		Lang:      constant.ScanLangName(constant.ParseLang(req.Lang).LowerString()),
		Tag:       req.Tag,
		Type:      constant.ScanRuleType(req.Type),
		Plugin:    req.Plugin,
		IsCustom:  req.IsCustom,
		Status:    constant.ScanRuleStatusEnabled,
		CreatedBy: session.User.Account,
		EditedBy:  session.User.Account,
	}

	return s.ruleSetStore.Create(ctx, ruleSet)
}

func (s *RuleSetService) GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error) {
	return s.ruleSetStore.GetByID(ctx, id)
}

func (s *RuleSetService) Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.RuleSetReqForUpdate) (*scanModel.RuleSet, error) {
	updates := map[string]interface{}{
		"editedBy":   session.User.Account,
		"editedDate": datetime.NowLocalTime(),
	}

	if req.Desc != nil {
		updates["desc"] = *req.Desc
	}
	if req.DescEn != nil {
		updates["descEn"] = *req.DescEn
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Lang != nil {
		updates["lang"] = constant.ScanLangName(constant.ParseLang(*req.Lang).LowerString())
	}
	if req.Tag != nil {
		updates["tag"] = *req.Tag
	}
	if req.Type != "" {
		updates["type"] = constant.ScanRuleType(req.Type)
	}
	if req.Plugin != "" {
		updates["plugin"] = req.Plugin
	}
	if req.Status != "" {
		updates["status"] = constant.ScanRuleStatus(req.Status)
	}

	return s.ruleSetStore.Update(ctx, id, updates)
}

func (s *RuleSetService) Delete(ctx context.Context, id uint) error {
	// 查询规则集详情
	ruleSet, err := s.ruleSetStore.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否为内置规则集
	if !ruleSet.IsCustom {
		return errors.WithI18nKey(nil, i18n.RuleSetBuiltinDeleteError)
	}

	// 检查规则集是否被扫描方案使用
	isUsed, err := s.ruleSetStore.IsUsedBySolution(ctx, id)
	if err != nil {
		return err
	}
	if isUsed {
		return errors.WithI18nKey(nil, i18n.RuleSetInUseError)
	}

	return s.ruleSetStore.Delete(ctx, id)
}

func (s *RuleSetService) List(ctx context.Context, filter scanModel.RuleSetReqFilter) (*database.PageResult[*scanModel.RuleSet], error) {
	return s.ruleSetStore.List(ctx, filter)
}

// GetBindRule 根据规则集ID获取关联的规则列表
func (s *RuleSetService) GetBindRule(ctx context.Context, ruleSetID uint, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error) {
	// 先获取规则集关联的规则ID列表
	ruleIDs, err := s.ruleSetStore.GetRuleIDList(ctx, ruleSetID)
	if err != nil {
		return nil, err
	}

	// 如果没有规则ID，返回空结果
	if len(ruleIDs) == 0 {
		return &database.PageResult[*scanModel.Rule]{
			Data:     []*scanModel.Rule{},
			Total:    0,
			Page:     filter.Page,
			PageSize: filter.Size,
		}, nil
	}

	// 调用 store 层获取规则列表
	return s.ruleStore.ListByRuleSetID(ctx, ruleIDs, filter)
}

// GetUnBindRule 根据规则集ID获取未关联的规则列表
func (s *RuleSetService) GetUnBindRule(ctx context.Context, ruleSetID uint, filter scanModel.RuleReqFilter) (*database.PageResult[*scanModel.Rule], error) {
	// 先获取规则集关联的规则ID列表
	ruleIDs, err := s.ruleSetStore.GetRuleIDList(ctx, ruleSetID)
	if err != nil {
		return nil, err
	}

	// 调用 store 层获取排除已绑定ID的规则列表
	return s.ruleStore.ExcludeByRuleSetID(ctx, ruleIDs, filter)
}

// BindRule 绑定规则到规则集
func (s *RuleSetService) BindRule(ctx context.Context, ruleSetID uint, req scanModel.RuleSetReqForBind) error {
	return s.ruleSetStore.BindRule(ctx, ruleSetID, req.IDs)
}

// UnBindRule 解绑规则与规则集
func (s *RuleSetService) UnBindRule(ctx context.Context, ruleSetID uint, req scanModel.RuleSetReqForBind) error {
	return s.ruleSetStore.UnBindRule(ctx, ruleSetID, req.IDs)
}
