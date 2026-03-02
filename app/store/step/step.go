package step

import (
	"context"
	"encoding/json"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

// Create creates a new step
func (s *Store) Create(ctx context.Context, step *stepModel.Step) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	dbStep := mapToDBModel(step)

	if err := db.Create(dbStep).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	step.ID = dbStep.ID
	return nil
}

// Update updates a step
func (s *Store) Update(ctx context.Context, step *stepModel.Step) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	dbStep := mapToDBModel(step)

	if err := db.Model(&Step{}).Where("id = ?", step.ID).Updates(dbStep).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	return nil
}

// GetByID gets step by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*stepModel.Step, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStep Step
	if err := db.Where("id = ?", id).First(&dbStep).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbStep), nil
}

// mapToDBModel converts business model to database model
func mapToDBModel(step *stepModel.Step) *Step {

	var dependsJSON string
	if step.Depends != nil {
		bytes, _ := json.Marshal(step.Depends)
		dependsJSON = string(bytes)
	}
	return &Step{
		ID:         step.ID,
		StageID:    step.StageID,
		Name:       step.Name,
		Code:       step.Code,
		SN:         step.SN,
		Image:      step.Image,
		Depends:    dependsJSON,
		GroupID:    step.GroupID,
		FailPolicy: step.FailPolicy,
		Started:    step.Started,
		Finished:   step.Finished,
		Duration:   step.Duration,
		Status:     string(step.Status),
		ErrorCode:  step.ErrorCode,
		ErrorLog:   step.ErrorLog,
	}
}

// mapToModel converts database model to business model
func mapToModel(s *Step) *stepModel.Step {

	var depends []string
	if s.Depends != "" {
		_ = json.Unmarshal([]byte(s.Depends), &depends)
	}
	return &stepModel.Step{
		ID:         s.ID,
		StageID:    s.StageID,
		Name:       s.Name,
		Code:       s.Code,
		SN:         s.SN,
		Image:      s.Image,
		Depends:    depends,
		GroupID:    s.GroupID,
		FailPolicy: s.FailPolicy,
		Started:    s.Started,
		Finished:   s.Finished,
		Duration:   s.Duration,
		Status:     constant.CIStatus(s.Status),
		ErrorCode:  s.ErrorCode,
		ErrorLog:   s.ErrorLog,
	}
}
