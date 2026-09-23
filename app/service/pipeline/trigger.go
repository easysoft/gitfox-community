// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"context"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/errors"
)

// CreateTrigger 创建触发
func (s *Service) CreateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	trigger.ID = 0
	trigger.EditedBy = ""
	trigger.EditedDate = nil

	if err := s.pipelineStore.CreateTrigger(ctx, trigger); err != nil {
		return err
	}

	// trigger.ID is populated by the store after insert.
	if err := s.SyncTriggerCronJobs(ctx, trigger.ID, trigger.PipelineID, trigger.Cron); err != nil {
		return errors.Newf("failed to sync cron jobs after creating trigger: %v", err)
	}

	return nil
}

// UpdateTrigger 更新触发
func (s *Service) UpdateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error {
	// Update DB first to avoid inconsistency: if the DB update fails we
	// haven't touched the scheduler yet.
	if err := s.pipelineStore.UpdateTrigger(ctx, trigger); err != nil {
		return err
	}

	// SyncTriggerCronJobs already removes all existing cron jobs for this
	// trigger before adding new ones — a single call is sufficient.
	if err := s.SyncTriggerCronJobs(ctx, trigger.ID, trigger.PipelineID, trigger.Cron); err != nil {
		return errors.Newf("failed to sync cron jobs after updating trigger: %v", err)
	}

	return nil
}

// GetTriggerByID 根据触发ID获取触发
func (s *Service) GetTriggerByID(ctx context.Context, id int64) (*pipelineModel.Trigger, error) {
	return s.pipelineStore.GetTriggerByID(ctx, id)
}

// ListTriggers 查询触发列表
func (s *Service) GetTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) ([]*pipelineModel.Trigger, error) {
	return s.pipelineStore.GetTriggers(ctx, filter)
}

// DeleteTrigger 删除触发
func (s *Service) DeleteTrigger(ctx context.Context, id int64) error {
	trigger, err := s.pipelineStore.GetTriggerByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from DB first. If this fails, the scheduler is untouched and
	// remains consistent with the database.
	if err := s.pipelineStore.DeleteTrigger(ctx, id); err != nil {
		return err
	}

	// Best-effort cleanup: remove cron jobs. Even if this fails, the orphan
	// jobs will simply fail when they try to look up the deleted trigger.
	if err := s.SyncTriggerCronJobs(ctx, trigger.ID, trigger.PipelineID, nil); err != nil {
		return errors.Newf("failed to remove cron jobs after deleting trigger: %v", err)
	}

	return nil
}
