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
		CreatedBy: session.User.Account,
		EditedBy:  session.User.Account,
	}

	return s.ruleSetStore.Create(ctx, ruleSet)
}

func (s *RuleSetService) GetByID(ctx context.Context, id uint) (*scanModel.RuleSet, error) {
	return s.ruleSetStore.GetByID(ctx, id)
}

func (s *RuleSetService) Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.RuleSetUpdateReq) (*scanModel.RuleSet, error) {
	// 将请求转换为 store 层的 RuleSet 结构
	ruleSet := &scanStore.RuleSet{EditedBy: session.User.Account}

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

func (s *RuleSetService) Delete(ctx context.Context, id uint) error {
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

func (s *RuleSetService) List(ctx context.Context, filter scanModel.RuleSetListReq) (*database.PageResult[*scanModel.RuleSet], error) {
	storeResult, err := s.ruleSetStore.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 转换 store 层数据为 model 层数据
	modelData := make([]*scanModel.RuleSet, 0, len(storeResult.Data))
	for _, item := range storeResult.Data {
		modelData = append(modelData, &scanModel.RuleSet{
			ID:          item.ID,
			Name:        item.Name,
			Desc:        item.Desc,
			DescEn:      item.DescrEn,
			Lang:        item.Lang,
			Tag:         item.Tag,
			IsCustom:    item.IsCustom,
			Status:      item.Status,
			CreatedBy:   item.CreatedBy,
			CreatedDate: item.CreatedDate,
			EditedBy:    item.EditedBy,
			EditedDate:  item.EditedDate,
			Deleted:     item.Deleted,
		})
	}

	return &database.PageResult[*scanModel.RuleSet]{
		Data:     modelData,
		Total:    storeResult.Total,
		Page:     storeResult.Page,
		PageSize: storeResult.PageSize,
	}, nil
}
