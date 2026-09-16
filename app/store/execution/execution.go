package execution

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/database/dbtx"
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
		PipelineID:   execution.PipelineID,
		Parent:       execution.Parent,
		Trigger:      execution.Trigger,
		Commit:       execution.Commit,
		CommitBefore: execution.CommitBefore,
		Ref:          execution.Ref,
		Params:       database.NoNull(paramsJSON),
		StartedDate:  execution.StartedDate,
		FinishedDate: execution.FinishedDate,
		Duration:     execution.Duration,
		Status:       string(execution.Status),
		Error:        execution.Error,
		Number:       execution.Number,
		Logs:         execution.Logs,
		CreatedBy:    execution.CreatedBy,
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
		PipelineID:   execution.PipelineID,
		Parent:       execution.Parent,
		Trigger:      execution.Trigger,
		Commit:       execution.Commit,
		CommitBefore: execution.CommitBefore,
		Ref:          execution.Ref,
		Params:       database.NoNull(paramsJSON),
		StartedDate:  execution.StartedDate,
		FinishedDate: execution.FinishedDate,
		Duration:     execution.Duration,
		Status:       string(execution.Status),
		Error:        execution.Error,
		Number:       execution.Number,
		Logs:         execution.Logs,
		CreatedBy:    execution.CreatedBy,
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
	if e.Params.String != "" {
		_ = json.Unmarshal([]byte(e.Params.String), &params)
	}

	return &executionModel.Execution{
		ID:           e.ID,
		PipelineID:   e.PipelineID,
		Parent:       e.Parent,
		Trigger:      e.Trigger,
		Commit:       e.Commit,
		CommitBefore: e.CommitBefore,
		Ref:          e.Ref,
		Params:       params,
		StartedDate:  e.StartedDate,
		FinishedDate: e.FinishedDate,
		Duration:     e.Duration,
		Status:       constant.CIStatus(e.Status),
		Error:        e.Error,
		Number:       e.Number,
		Logs:         e.Logs,
		CreatedBy:    e.CreatedBy,
		CreatedDate:  e.CreatedDate,
		EditedDate:   e.EditedDate,
	}
}

func populateDuration(execution *executionModel.Execution) {
	if execution == nil || execution.StartedDate == nil || execution.FinishedDate == nil {
		return
	}
	duration := time.Time(*execution.FinishedDate).Sub(time.Time(*execution.StartedDate))
	execution.Duration = uint(math.Ceil(duration.Seconds()))
}

// ListByPr gets executions by PR with pipeline name from ops_pipeline
func (s *Store) ListByPr(ctx context.Context, filter executionModel.ExecutionInfoReqFilter) ([]*executionModel.ExecutionByPr, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var results []*executionModel.ExecutionByPr
	query := db.Table("ops_pipeline_executions e").
		Select("e.id, e.PipelineID, p.name, e.trigger, e.commit, e.ref, e.status").
		Joins("JOIN ops_pipeline p ON e.PipelineID = p.id").
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
