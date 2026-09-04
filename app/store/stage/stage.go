package stage

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	stepStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/step"
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
		Status:       string(stage.Status),
		ErrorType:    stage.ErrorType,
		ErrorCode:    stage.ErrorCode,
		ErrorLog:     database.NullIfEmpty(stage.ErrorLog),
	}

	if err := db.Create(dbStage).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	stage.ID = dbStage.ID
	return nil
}

// Update updates a stage
func (s *Store) Update(ctx context.Context, stage *stageModel.Stage) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)
	populateDuration(stage)

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
		Status:       string(stage.Status),
		ErrorType:    stage.ErrorType,
		ErrorCode:    stage.ErrorCode,
		ErrorLog:     database.NullIfEmpty(stage.ErrorLog),
	}

	// Omit scheduling metadata fields that are set once at creation time.
	// Runners must not overwrite these via UpdateStage/UpdateStep callbacks.
	if err := db.Model(&Stage{}).Where("id = ?", stage.ID).
		Omit("code", "type", "labels", "depends", "run_policy", "custom_policy", "fail_policy").
		Updates(dbStage).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

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

// GetByIDWithSteps gets stage by ID
func (s *Store) GetByIDWithSteps(ctx context.Context, id uint) (*stageModel.StageWithSteps, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStage Stage
	if err := db.Where("id = ?", id).First(&dbStage).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	stage := mapToModel(&dbStage)

	var dbSteps []stepStore.Step
	if err := db.Where("stageID = ?", id).Order("sn ASC").Find(&dbSteps).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	steps := make([]*stepModel.Step, len(dbSteps))
	for _, dbStep := range dbSteps {
		var depends []string
		if dbStep.Depends != "" {
			_ = json.Unmarshal([]byte(dbStep.Depends), &depends)
		}
		step := &stepModel.Step{
			ID:         dbStep.ID,
			StageID:    dbStep.StageID,
			Name:       dbStep.Name,
			Code:       dbStep.Code,
			SN:         dbStep.SN,
			Image:      dbStep.Image,
			Depends:    depends,
			GroupID:    dbStep.GroupID,
			FailPolicy: dbStep.FailPolicy,
			Started:    dbStep.Started,
			Finished:   dbStep.Finished,
			Duration:   dbStep.Duration,
			Status:     constant.CIStatus(dbStep.Status),
			ErrorCode:  dbStep.ErrorCode,
			ErrorLog:   dbStep.ErrorLog,
		}
		steps = append(steps, step)
	}

	result := stageModel.StageWithSteps{
		Stage: *stage,
		Steps: steps,
	}
	return &result, nil
}

// FindByNumber returns a stage by execution ID and stage number (sn)
func (s *Store) FindByNumber(ctx context.Context, executionID int64, sn int) (*stageModel.Stage, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStage Stage
	if err := db.Where("executionID = ? AND sn = ?", executionID, sn).First(&dbStage).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbStage), nil
}

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
		Status:       constant.CIStatus(s.Status),
		ErrorType:    s.ErrorType,
		ErrorCode:    s.ErrorCode,
		ErrorLog:     s.ErrorLog.String,
	}
}

// ListIncomplete returns a stage list where the stage is incomplete (pending or running)
func (s *Store) ListIncomplete(ctx context.Context) ([]*stageModel.Stage, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStages []Stage
	if err := db.Where("status IN ?", []string{"pending", "running"}).Order("sn ASC").Find(&dbStages).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	stages := make([]*stageModel.Stage, len(dbStages))
	for i := range dbStages {
		stages[i] = mapToModel(&dbStages[i])
	}
	return stages, nil
}

// List returns a list of stages corresponding to an execution ID
func (s *Store) List(ctx context.Context, executionID uint) ([]*stageModel.Stage, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbStages []Stage
	if err := db.Where("executionID = ?", executionID).Order("sn ASC").Find(&dbStages).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	stages := make([]*stageModel.Stage, len(dbStages))
	for i := range dbStages {
		stages[i] = mapToModel(&dbStages[i])
	}
	return stages, nil
}

// ListWithSteps returns a stage list with steps for an execution
func (s *Store) ListWithSteps(ctx context.Context, executionID uint) ([]*stageModel.StageWithSteps, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	// Get all stages for the execution
	var dbStages []Stage
	if err := db.Where("executionID = ?", executionID).Order("sn ASC").Find(&dbStages).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	if len(dbStages) == 0 {
		return []*stageModel.StageWithSteps{}, nil
	}

	// Collect stage IDs
	stageIDs := make([]uint, len(dbStages))
	for i, stage := range dbStages {
		stageIDs[i] = stage.ID
	}

	// Get all steps for these stages
	var dbSteps []stepStore.Step
	if err := db.Where("stageID IN ?", stageIDs).Order("sn ASC").Find(&dbSteps).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// Group steps by stageID
	stepsByStageID := make(map[uint][]*stepModel.Step)
	for _, dbStep := range dbSteps {
		var depends []string
		if dbStep.Depends != "" {
			_ = json.Unmarshal([]byte(dbStep.Depends), &depends)
		}
		step := &stepModel.Step{
			ID:         dbStep.ID,
			StageID:    dbStep.StageID,
			Name:       dbStep.Name,
			Code:       dbStep.Code,
			SN:         dbStep.SN,
			Image:      dbStep.Image,
			Depends:    depends,
			GroupID:    dbStep.GroupID,
			FailPolicy: dbStep.FailPolicy,
			Started:    dbStep.Started,
			Finished:   dbStep.Finished,
			Duration:   dbStep.Duration,
			Status:     constant.CIStatus(dbStep.Status),
			ErrorCode:  dbStep.ErrorCode,
			ErrorLog:   dbStep.ErrorLog,
		}
		stepsByStageID[dbStep.StageID] = append(stepsByStageID[dbStep.StageID], step)
	}

	// Build result
	result := make([]*stageModel.StageWithSteps, len(dbStages))
	for i := range dbStages {
		stage := mapToModel(&dbStages[i])
		steps := stepsByStageID[dbStages[i].ID]
		if steps == nil {
			steps = []*stepModel.Step{}
		}
		result[i] = &stageModel.StageWithSteps{
			Stage: *stage,
			Steps: steps,
		}
	}

	return result, nil
}

func populateDuration(stage *stageModel.Stage) {
	if stage == nil || stage.Started == nil || stage.Finished == nil {
		return
	}
	duration := time.Time(*stage.Finished).Sub(time.Time(*stage.Started))
	stage.Duration = uint(math.Ceil(duration.Seconds()))
}
