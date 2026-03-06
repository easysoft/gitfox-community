package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

var _ IScanCondition = (*ScanConditionService)(nil)

type ScanConditionService struct {
	conditionStore scanStore.IScanCondition
}

func NewScanConditionService(conditionStore scanStore.IScanCondition) *ScanConditionService {
	return &ScanConditionService{conditionStore: conditionStore}
}

func (s *ScanConditionService) BatchCreate(ctx context.Context, session *authModel.Session, planID uint, triggerID uint, repoID uint, req scanModel.ScanConditionReqForBatchCreate) error {
	conditions := make([]*scanStore.ScanPlanCondition, 0, len(req.Conditions))
	for _, item := range req.Conditions {
		conditions = append(conditions, &scanStore.ScanPlanCondition{
			PlanID:    uint64(planID),
			TriggerID: uint64(triggerID),
			RepoID:    uint64(repoID),
			Priority:  constant.ScanRulePriority(item.Priority),
			Type:      constant.ScanRuleType(item.Type),
			Unit:      constant.ScanRuleUnit(item.Unit),
			Threshold: item.Threshold,
			CreatedBy: session.User.Account,
			EditedBy:  session.User.Account,
		})
	}
	return s.conditionStore.BatchCreate(ctx, conditions)
}

func (s *ScanConditionService) BatchDelete(ctx context.Context, repoID uint, planID uint, triggerID uint, req scanModel.ScanConditionReqForBatchDelete) error {
	return s.conditionStore.BatchDelete(ctx, repoID, planID, triggerID, req.IDs)
}

func (s *ScanConditionService) List(ctx context.Context, repoID uint, planID uint, triggerID uint) ([]*scanModel.ScanCondition, error) {
	conditions, err := s.conditionStore.List(ctx, triggerID)
	if err != nil {
		return nil, err
	}
	filtered := make([]*scanModel.ScanCondition, 0, len(conditions))
	for _, c := range conditions {
		if uint(c.RepoID) == repoID && uint(c.PlanID) == planID {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}
