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

var _ ISolution = (*SolutionService)(nil)

type SolutionService struct {
	solutionStore scanStore.ISolution
}

func NewSolutionService(solutionStore scanStore.ISolution) *SolutionService {
	return &SolutionService{solutionStore: solutionStore}
}

func (s *SolutionService) Create(ctx context.Context, session *authModel.Session, req scanModel.SolutionReqForCreate) (*scanModel.Solution, error) {
	existing, err := s.solutionStore.Find(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.WithI18nKey(nil, i18n.SolutionAlreadyExists)
	}

	solution := &scanStore.SolutionRules{
		Name:      req.Name,
		Desc:      req.Desc,
		DescEn:    req.DescEn,
		Tag:       req.Tag,
		IsCustom:  true,
		Status:    req.Status,
		CreatedBy: session.User.Account,
		EditedBy:  session.User.Account,
	}

	return s.solutionStore.Create(ctx, solution)
}

func (s *SolutionService) GetByID(ctx context.Context, id uint) (*scanModel.Solution, error) {
	solution, err := s.getSolutionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 聚合关联的规则集信息
	rulesets, err := s.solutionStore.ListRulesets(ctx, id)
	if err != nil {
		return nil, err
	}
	solution.RuleSets = rulesets
	solution.RuleSetsCount = int64(len(rulesets))

	// 聚合 RulesCount、Plugins、Langs
	pluginMap := make(map[string]struct{})
	langMap := make(map[string]struct{})
	for _, rs := range rulesets {
		solution.RulesCount += rs.RulesCount
		if rs.Plugin != "" {
			pluginMap[rs.Plugin] = struct{}{}
		}
		if rs.Lang != "" {
			langMap[string(rs.Lang)] = struct{}{}
		}
	}
	for plugin := range pluginMap {
		solution.Plugins = append(solution.Plugins, plugin)
	}
	for lang := range langMap {
		solution.Langs = append(solution.Langs, lang)
	}

	return solution, nil
}

func (s *SolutionService) Update(ctx context.Context, session *authModel.Session, id uint, req scanModel.SolutionReqForUpdate) (*scanModel.Solution, error) {
	solution := &scanStore.SolutionRules{
		ID:       uint64(id),
		EditedBy: session.User.Account,
	}
	if req.Name != "" {
		solution.Name = req.Name
	}
	if req.Desc != "" {
		solution.Desc = req.Desc
	}
	if req.DescEn != "" {
		solution.DescEn = req.DescEn
	}
	if req.Tag != "" {
		solution.Tag = req.Tag
	}

	if req.Status != "" {
		solution.Status = req.Status
	}

	if err := s.solutionStore.Update(ctx, solution); err != nil {
		return nil, err
	}
	return s.getSolutionByID(ctx, id)
}

func (s *SolutionService) Delete(ctx context.Context, id uint) error {
	obj, err := s.getSolutionByID(ctx, id)
	if err != nil {
		return err
	}
	if !obj.IsCustom {
		return errors.WithI18nKey(nil, i18n.SolutionBuiltinDeleteError)
	}
	return s.solutionStore.Delete(ctx, id)
}

func (s *SolutionService) List(ctx context.Context, filter scanModel.SolutionReqFilter) (*database.PageResult[*scanModel.Solution], error) {
	return s.solutionStore.List(ctx, filter)
}

// UpdateStatus 切换扫描方案的状态（enabled ↔ disabled），适用于 /status 和 /draft 两个接口
func (s *SolutionService) UpdateStatus(ctx context.Context, session *authModel.Session, id uint) error {
	obj, err := s.getSolutionByID(ctx, id)
	if err != nil {
		return err
	}

	newStatus := constant.ScanRuleStatusEnabled
	if obj.Status == constant.ScanRuleStatusEnabled {
		newStatus = constant.ScanRuleStatusDisabled
	}

	solution := &scanStore.SolutionRules{
		ID:       uint64(id),
		Status:   newStatus,
		EditedBy: session.User.Account,
	}
	return s.solutionStore.Update(ctx, solution)
}

func (s *SolutionService) ListRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error) {
	if _, err := s.getSolutionByID(ctx, solutionID); err != nil {
		return nil, err
	}
	return s.solutionStore.ListRulesets(ctx, solutionID)
}

func (s *SolutionService) ListUnbindRulesets(ctx context.Context, solutionID uint) ([]*scanModel.RuleSet, error) {
	if _, err := s.getSolutionByID(ctx, solutionID); err != nil {
		return nil, err
	}
	return s.solutionStore.ListUnbindRulesets(ctx, solutionID)
}

func (s *SolutionService) BindRulesets(ctx context.Context, solutionID uint, req scanModel.SolutionReqForBind) error {
	if len(req.IDs) == 0 {
		return errors.WithI18nKey(errors.New("rulesetIDList is required"), i18n.SolutionRulesetIDsRequired)
	}
	if _, err := s.getSolutionByID(ctx, solutionID); err != nil {
		return err
	}
	return s.solutionStore.BindRulesets(ctx, solutionID, req.IDs)
}

func (s *SolutionService) UnbindRulesets(ctx context.Context, solutionID uint, req scanModel.SolutionReqForBind) error {
	if len(req.IDs) == 0 {
		return errors.WithI18nKey(errors.New("rulesetIDList is required"), i18n.SolutionRulesetIDsRequired)
	}
	if _, err := s.getSolutionByID(ctx, solutionID); err != nil {
		return err
	}
	return s.solutionStore.UnbindRulesets(ctx, solutionID, req.IDs)
}

// getSolutionByID 加载扫描方案，并把底层 not found 转成 SolutionNotFound.
func (s *SolutionService) getSolutionByID(ctx context.Context, id uint) (*scanModel.Solution, error) {
	solution, err := s.solutionStore.GetByID(ctx, id)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.SolutionNotFound)
		}
		return nil, err
	}
	return solution, nil
}
