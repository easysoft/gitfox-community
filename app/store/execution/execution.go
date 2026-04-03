package execution

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

// Create creates a new execution
func (s *Store) Create(ctx context.Context, execution *executionModel.Execution) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	// Convert params map to JSON string
	var paramsJSON string
	if execution.Params != nil {
		bytes, err := json.Marshal(execution.Params)
		if err != nil {
			return err
		}
		paramsJSON = string(bytes)
	}

	dbExecution := &Execution{
		PipelineID: execution.PipelineID,
		Parent:     execution.Parent,
		Trigger:    execution.Trigger,
		Commit:     execution.Commit,
		Ref:        execution.Ref,
		Params:     paramsJSON,
		Started:    execution.Started,
		Finished:   execution.Finished,
		Duration:   execution.Duration,
		Status:     string(execution.Status),
		Error:      execution.Error,
		CreatedBy:  execution.CreatedBy,
	}

	if err := db.Create(dbExecution).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	execution.ID = dbExecution.ID
	execution.CreatedDate = dbExecution.CreatedDate
	execution.EditedDate = dbExecution.EditedDate
	return nil
}

// Update updates an execution
func (s *Store) Update(ctx context.Context, execution *executionModel.Execution) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)
	populateDuration(execution)

	var paramsJSON string
	if execution.Params != nil {
		bytes, err := json.Marshal(execution.Params)
		if err != nil {
			return err
		}
		paramsJSON = string(bytes)
	}

	dbExecution := &Execution{
		PipelineID: execution.PipelineID,
		Parent:     execution.Parent,
		Trigger:    execution.Trigger,
		Commit:     execution.Commit,
		Ref:        execution.Ref,
		Params:     paramsJSON,
		Started:    execution.Started,
		Finished:   execution.Finished,
		Duration:   execution.Duration,
		Status:     string(execution.Status),
		Error:      execution.Error,
		CreatedBy:  execution.CreatedBy,
	}

	if err := db.Model(&Execution{}).Where("id = ?", execution.ID).Updates(dbExecution).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	return nil
}

// GetByID gets execution by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*executionModel.Execution, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbExecution Execution
	if err := db.Where("id = ?", id).First(&dbExecution).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbExecution), nil
}

// mapToModel converts database model to business model
func mapToModel(e *Execution) *executionModel.Execution {
	var params map[string]string
	if e.Params != "" {
		_ = json.Unmarshal([]byte(e.Params), &params)
	}

	return &executionModel.Execution{
		ID:          e.ID,
		PipelineID:  e.PipelineID,
		Parent:      e.Parent,
		Trigger:     e.Trigger,
		Commit:      e.Commit,
		Ref:         e.Ref,
		Params:      params,
		Started:     e.Started,
		Finished:    e.Finished,
		Duration:    e.Duration,
		Status:      constant.CIStatus(e.Status),
		Error:       e.Error,
		CreatedBy:   e.CreatedBy,
		CreatedDate: e.CreatedDate,
		EditedDate:  e.EditedDate,
	}
}

func populateDuration(execution *executionModel.Execution) {
	if execution == nil || execution.Started == nil || execution.Finished == nil {
		return
	}
	duration := time.Time(*execution.Finished).Sub(time.Time(*execution.Started))
	execution.Duration = uint(math.Ceil(duration.Seconds()))
}

// ListByPr gets executions by PR with pipeline name from ops_pipeline
func (s *Store) ListByPr(ctx context.Context, filter executionModel.ExecutionInfoReqFilter) ([]*executionModel.ExecutionByPr, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var results []*executionModel.ExecutionByPr
	query := db.Table("ops_pipeline_executions e").
		Select("e.id, e.pipelineID, p.name, e.trigger, e.commit, e.ref, e.status").
		Joins("JOIN ops_pipeline p ON e.pipelineID = p.id").
		Where("p.deleted = 0")

	if filter.Commit != "" {
		query = query.Where("e.commit = ?", filter.Commit)
	}
	if filter.Ref != "" {
		query = query.Where("e.ref = ?", filter.Ref)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return results, nil
}
