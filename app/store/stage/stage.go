package stage

import (
	"context"
	"encoding/json"

	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

// Create creates a new stage
func (s *Store) Create(ctx context.Context, stage *stageModel.Stage) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	// Convert maps/slices to JSON strings
	var labelsJSON, dependsJSON string
	if stage.Labels != nil {
		bytes, err := json.Marshal(stage.Labels)
		if err != nil {
			return err
		}
		labelsJSON = string(bytes)
	}
	if stage.Depends != nil {
		bytes, err := json.Marshal(stage.Depends)
		if err != nil {
			return err
		}
		dependsJSON = string(bytes)
	}

	dbStage := &Stage{
		RepoID:       stage.RepoID,
		ExecutionID:  stage.ExecutionID,
		Parent:       stage.Parent,
		Name:         stage.Name,
		Code:         stage.Code,
		SN:           stage.SN,
		Type:         stage.Type,
		Host:         stage.Host,
		OS:           stage.OS,
		Arch:         stage.Arch,
		Labels:       labelsJSON,
		Depends:      dependsJSON,
		RunPolicy:    stage.RunPolicy,
		CustomPolicy: stage.CustomPolicy,
		FailPolicy:   stage.FailPolicy,
		Started:      stage.Started,
		Finished:     stage.Finished,
		Duration:     stage.Duration,
		Status:       stage.Status,
		ErrorType:    stage.ErrorType,
		ErrorCode:    stage.ErrorCode,
		ErrorLog:     stage.ErrorLog,
	}

	if err := db.Create(dbStage).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	stage.ID = dbStage.ID
	return nil
}

// GetByID gets stage by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*stageModel.Stage, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStage Stage
	if err := db.Where("id = ?", id).First(&dbStage).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbStage), nil
}

// mapToModel converts database model to business model
func mapToModel(s *Stage) *stageModel.Stage {
	var labels map[string]string
	var depends []string

	if s.Labels != "" {
		_ = json.Unmarshal([]byte(s.Labels), &labels)
	}
	if s.Depends != "" {
		_ = json.Unmarshal([]byte(s.Depends), &depends)
	}

	return &stageModel.Stage{
		ID:           s.ID,
		RepoID:       s.RepoID,
		ExecutionID:  s.ExecutionID,
		Parent:       s.Parent,
		Name:         s.Name,
		Code:         s.Code,
		SN:           s.SN,
		Type:         s.Type,
		Host:         s.Host,
		OS:           s.OS,
		Arch:         s.Arch,
		Labels:       labels,
		Depends:      depends,
		RunPolicy:    s.RunPolicy,
		CustomPolicy: s.CustomPolicy,
		FailPolicy:   s.FailPolicy,
		Started:      s.Started,
		Finished:     s.Finished,
		Duration:     s.Duration,
		Status:       s.Status,
		ErrorType:    s.ErrorType,
		ErrorCode:    s.ErrorCode,
		ErrorLog:     s.ErrorLog,
	}
}
