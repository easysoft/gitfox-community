package scan

import (
	"context"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ IScanTask = (*ScanTaskStore)(nil)

type ScanTaskStore struct {
	db *gorm.DB
}

func NewScanTaskStore(db *gorm.DB) *ScanTaskStore {
	return &ScanTaskStore{db: db}
}

// Create 创建扫描任务
func (s *ScanTaskStore) Create(ctx context.Context, task *ScanTask) (*scanModel.ScanTask, error) {
	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTaskToModel(*task), nil
}

// GetByID 根据ID查询扫描任务
func (s *ScanTaskStore) GetByID(ctx context.Context, id uint) (*scanModel.ScanTask, error) {
	var task ScanTask
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false).First(&task).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTaskToModel(task), nil
}

// FindByExecID 根据执行ID查询扫描任务
func (s *ScanTaskStore) FindByExecID(ctx context.Context, execID uint) (*scanModel.ScanTask, error) {
	var task ScanTask
	err := s.db.WithContext(ctx).Where("executionID = ? AND deleted = ?", execID, false).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTaskToModel(task), nil
}

func (s *ScanTaskStore) baseTaskInfoQuery(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Table("ops_scan_tasks").
		Select("ops_scan_tasks.*, "+
			"ops_scan_plans.name as planName, ops_scan_plans.repoID, "+
			"ops_repo.name as repoName, "+
			"t.name as triggerName, "+
			"e.id as executionNumber, e.ref as executionRef, "+
			"e.status as executionStatus, e.started as executionStarted, e.finished as executionFinished").
		Joins("INNER JOIN ops_scan_plans ON ops_scan_tasks.planID = ops_scan_plans.id").
		Joins("INNER JOIN ops_repo ON ops_scan_plans.repoID = ops_repo.id").
		Joins("INNER JOIN ops_pipeline_executions e ON ops_scan_tasks.executionID = e.id").
		Joins("LEFT JOIN ops_scan_triggers t ON ops_scan_tasks.triggerID = t.id").
		Where("ops_scan_tasks.deleted = ?", false)
}

// GetTaskInfo 查询扫描任务详情（含关联信息）
func (s *ScanTaskStore) GetTaskInfo(ctx context.Context, taskID uint) (*scanModel.ScanTaskInfo, error) {
	var dst taskJoin
	if err := s.baseTaskInfoQuery(ctx).Where("ops_scan_tasks.id = ?", taskID).Take(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapTaskJoinToScanTaskInfo(&dst), nil
}

// List 查询指定计划的扫描任务列表
func (s *ScanTaskStore) List(ctx context.Context, planID uint) ([]*scanModel.ScanTask, error) {
	var tasks []*ScanTask
	if err := s.db.WithContext(ctx).
		Where("planID = ? AND deleted = ?", planID, false).
		Order("createdDate DESC").
		Find(&tasks).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapScanTaskListToModel(tasks), nil
}

// ListTaskInfos 查询扫描任务详情列表（含分页和过滤）
func (s *ScanTaskStore) ListTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (*database.PageResult[*scanModel.ScanTaskInfo], error) {
	stmt := s.baseTaskInfoQuery(ctx)
	stmt = applyTaskQueryFilter(stmt, filter)

	var taskInfos []*taskJoin
	pageResult, err := database.Paginate(ctx, stmt, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      normalizeScanTaskSortField(filter.Sort),
		Order:     normalizeSortOrder(filter.Order, "desc"),
		NeedTotal: true,
	}, &taskInfos)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	modelData := make([]*scanModel.ScanTaskInfo, 0, len(pageResult.Data))
	for _, item := range pageResult.Data {
		modelData = append(modelData, mapTaskJoinToScanTaskInfo(item))
	}

	return &database.PageResult[*scanModel.ScanTaskInfo]{
		Data:     modelData,
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
	}, nil
}

// ListBatch 批量查询扫描任务详情
func (s *ScanTaskStore) ListBatch(ctx context.Context, repoID int64, batchSize int, fn func(task *scanModel.ScanTaskInfo) error) error {
	var results []*taskJoin
	stmt := s.baseTaskInfoQuery(ctx)
	if repoID > 0 {
		stmt = stmt.Where("ops_scan_plans.repoID = ?", repoID)
	}
	res := stmt.FindInBatches(&results, batchSize, func(tx *gorm.DB, batch int) error {
		for _, result := range results {
			if err := fn(mapTaskJoinToScanTaskInfo(result)); err != nil {
				return err
			}
		}
		return nil
	})
	if res.Error != nil {
		return database.ProcessSQLErrorf(ctx, res.Error)
	}
	return nil
}

// CountTaskInfos 统计扫描任务详情数量
func (s *ScanTaskStore) CountTaskInfos(ctx context.Context, filter scanModel.ScanTaskInfoReqFilter) (int64, error) {
	stmt := s.baseTaskInfoQuery(ctx)
	stmt = applyTaskQueryFilter(stmt, filter)

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	return count, nil
}

// Update 更新扫描任务
func (s *ScanTaskStore) Update(ctx context.Context, task *ScanTask) error {
	if err := s.db.WithContext(ctx).Updates(task).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateResult 更新扫描任务结果
func (s *ScanTaskStore) UpdateResult(ctx context.Context, taskID uint, result constant.ScanTaskResult) error {
	if err := s.db.WithContext(ctx).Model(&ScanTask{}).
		Where("id = ?", taskID).
		Update("result", result).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateIssueNumber 更新扫描任务关联的问题编号
func (s *ScanTaskStore) UpdateIssueNumber(ctx context.Context, taskID uint, issueNumber int64) error {
	if err := s.db.WithContext(ctx).Model(&ScanTask{}).
		Where("id = ?", taskID).
		Update("issueNumber", issueNumber).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// Delete 软删除扫描任务
func (s *ScanTaskStore) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&ScanTask{}).
		Where("id = ?", id).Update("deleted", true).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// NextPlanNumber 获取指定计划的下一个任务序号
func (s *ScanTaskStore) NextPlanNumber(ctx context.Context, planID uint64) (int64, error) {
	var maxNum *int64
	if err := s.db.WithContext(ctx).Model(&ScanTask{}).
		Where("planID = ? AND deleted = ?", planID, false).
		Select("COALESCE(MAX(planNumber), 0)").Scan(&maxNum).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	if maxNum == nil {
		return 1, nil
	}
	return *maxNum + 1, nil
}

// NextRepoNumber 获取指定仓库的下一个扫描任务序号
func (s *ScanTaskStore) NextRepoNumber(ctx context.Context, repoID int64) (int64, error) {
	var maxNum *int64
	if err := s.db.WithContext(ctx).Table("ops_scan_tasks").
		Joins("INNER JOIN ops_scan_plans ON ops_scan_tasks.planID = ops_scan_plans.id").
		Where("ops_scan_plans.repoID = ? AND ops_scan_tasks.deleted = ?", repoID, false).
		Select("COALESCE(MAX(ops_scan_tasks.repoNumber), 0)").Scan(&maxNum).Error; err != nil {
		return 0, database.ProcessSQLErrorf(ctx, err)
	}
	if maxNum == nil {
		return 1, nil
	}
	return *maxNum + 1, nil
}

// ==================== 映射函数 ====================

func mapScanTaskToModel(data ScanTask) *scanModel.ScanTask {
	var triggerID *int64
	if data.TriggerID != nil {
		id := int64(*data.TriggerID)
		triggerID = &id
	}
	return &scanModel.ScanTask{
		ID:          int64(data.ID),
		PlanID:      int64(data.PlanID),
		ExecutionID: int64(data.ExecutionID),
		RepoNumber:  data.RepoNumber,
		PlanNumber:  data.PlanNumber,
		TriggerId:   triggerID,
		TriggerType: data.TriggerType,
		Status:      data.Status,
		Result:      data.Result,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedDate:  data.EditedDate,
		ScanType:    data.ScanType,
	}
}

func mapScanTaskListToModel(data []*ScanTask) []*scanModel.ScanTask {
	result := make([]*scanModel.ScanTask, 0, len(data))
	for _, item := range data {
		result = append(result, mapScanTaskToModel(*item))
	}
	return result
}

func mapTaskJoinToScanTaskInfo(in *taskJoin) *scanModel.ScanTaskInfo {
	info := &scanModel.ScanTaskInfo{
		ID:              int64(in.ID),
		RepoID:          in.RepoId,
		RepoName:        in.RepoName,
		PlanID:          int64(in.PlanID),
		PlanName:        in.PlanName,
		PipelineName:    "_codescan",
		ExecutionNumber: in.ExecutionNumber,
		ExecutionID:     int64(in.ExecutionID),
		ExecutionRef:    in.ExecutionRef,
		RepoNumber:      in.RepoNumber,
		PlanNumber:      in.PlanNumber,
		Result:          in.Result,
		ScanType:        in.ScanType,
		IssueNumber:     in.IssueNumber,
		CreatedBy:       in.CreatedBy,
		CreatedDate:     in.CreatedDate,
		EditedDate:      in.EditedDate,
	}

	// 转换执行时间为 unix 毫秒
	if in.ExecutionStarted != nil {
		info.Started = time.Time(*in.ExecutionStarted).UnixMilli()
	}
	if in.ExecutionFinished != nil {
		info.Finished = time.Time(*in.ExecutionFinished).UnixMilli()
	}

	// 计算执行耗时（秒）
	if info.Finished > info.Started {
		info.Cost = (info.Finished - info.Started) / 1000
	}

	// 设置触发器信息
	if in.TriggerType == constant.ScanPlanTriggerTypeAction || in.TriggerType == constant.ScanPlanTriggerTypeCron {
		info.Trigger = &scanModel.ScanTaskTrigger{
			Type: in.TriggerType,
		}
		if in.TriggerID != nil {
			info.Trigger.ID = int64(*in.TriggerID)
			info.Trigger.Name = in.TriggerName
		}
	} else {
		info.Trigger = &scanModel.ScanTaskTrigger{
			Type: constant.ScanPlanTriggerTypeManual,
		}
	}

	// 全量扫描 + 动作触发 → 检查是否应转为增量扫描
	if in.ScanType == constant.ScanTypeFull && in.TriggerType == constant.ScanPlanTriggerTypeAction {
		if strings.Contains(in.ExecutionRef, "refs/pullreq/") {
			info.ScanType = constant.ScanTypeIncremental
		}
	}

	// 根据执行状态转换任务状态
	info.Status = convertExecutionTaskStatus(in.ExecutionStatus)

	return info
}

func convertExecutionTaskStatus(in constant.CIStatus) constant.ScanTaskStatus {
	switch in {
	case constant.CIStatusSuccess:
		return constant.ScanTaskStatusSuccess
	case constant.CIStatusFailure, constant.CIStatusKilled, constant.CIStatusError:
		return constant.ScanTaskStatusFailed
	default:
		return constant.ScanTaskStatusInProgress
	}
}

// ==================== 查询过滤 ====================

func applyTaskQueryFilter(stmt *gorm.DB, filter scanModel.ScanTaskInfoReqFilter) *gorm.DB {
	if filter.TaskID != 0 {
		stmt = stmt.Where("ops_scan_tasks.id = ?", filter.TaskID)
	}
	if filter.RepoID != 0 {
		stmt = stmt.Where("ops_scan_plans.repoID = ?", filter.RepoID)
	}
	if filter.PlanID != 0 {
		stmt = stmt.Where("ops_scan_tasks.planID = ?", filter.PlanID)
	}
	if filter.TriggerID != 0 {
		stmt = stmt.Where("ops_scan_tasks.triggerID = ?", filter.TriggerID)
	}
	if filter.Branch != "" {
		stmt = stmt.Where("e.ref = ?", filter.Branch)
	}
	if filter.Status != "" {
		switch filter.Status {
		case constant.ScanTaskStatusSuccess:
			stmt = stmt.Where("e.status = ?", constant.CIStatusSuccess)
		case constant.ScanTaskStatusFailed:
			stmt = stmt.Where("e.status IN ?", []constant.CIStatus{
				constant.CIStatusError, constant.CIStatusFailure, constant.CIStatusKilled,
			})
		case constant.ScanTaskStatusInProgress:
			stmt = stmt.Where("e.status NOT IN ?", []constant.CIStatus{
				constant.CIStatusSuccess, constant.CIStatusError, constant.CIStatusFailure, constant.CIStatusKilled,
			})
		}
	}
	if filter.Result != "" {
		if filter.Result == constant.ScanTaskResultNull {
			stmt = stmt.Where("ops_scan_tasks.result IS NULL")
		} else {
			stmt = stmt.Where("ops_scan_tasks.result = ?", filter.Result)
		}
	}
	if filter.CreateDate != "" {
		stmt = stmt.Where("DATE(ops_scan_tasks.createdDate) = ?", filter.CreateDate[:10])
	}
	return stmt
}

func normalizeScanTaskSortField(sort string) string {
	switch strings.TrimSpace(sort) {
	case "started":
		return "e.started"
	case "finished":
		return "e.finished"
	case "repo_id":
		return "ops_scan_plans.repoID"
	case "status":
		return "e.status"
	case "result":
		return "ops_scan_tasks.result"
	default:
		return "ops_scan_tasks.id"
	}
}
