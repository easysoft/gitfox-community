package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	runnerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database/dbtx"
)

// Create creates a new runner
func (s *Store) Create(ctx context.Context, runner *runnerModel.Runner) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	// Convert labels slice to string
	var labels string
	if len(runner.Labels) > 0 {
		labels = fmt.Sprintf(",%s,", strings.Join(runner.Labels, ","))
	}

	dbRunner := &Runner{
		Name:      runner.Name,
		Desc:      runner.Desc,
		Version:   runner.Version,
		Hostname:  runner.Hostname,
		IP:        runner.IP,
		OS:        runner.OS,
		Arch:      runner.Arch,
		Runtime:   runner.Runtime,
		Labels:    labels,
		Token:     runner.Token,
		HeartBeat: runner.HeartBeat,
		Online:    string(runner.Online),
		Status:    string(runner.Status),
		CreatedBy: runner.CreatedBy,
		EditedBy:  runner.EditedBy,
		IsDefault: runner.IsDefault,
	}

	if err := db.Create(dbRunner).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	runner.ID = dbRunner.ID
	runner.CreatedDate = dbRunner.CreatedDate
	runner.EditedDate = dbRunner.EditedDate
	return nil
}

// Update updates a runner
func (s *Store) Update(ctx context.Context, runner *runnerModel.Runner) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var labelsJSON string
	if runner.Labels != nil {
		bytes, err := json.Marshal(runner.Labels)
		if err != nil {
			return err
		}
		labelsJSON = string(bytes)
	}

	dbRunner := &Runner{
		Name:      runner.Name,
		Desc:      runner.Desc,
		Version:   runner.Version,
		Hostname:  runner.Hostname,
		IP:        runner.IP,
		OS:        runner.OS,
		Arch:      runner.Arch,
		Runtime:   runner.Runtime,
		Labels:    labelsJSON,
		Token:     runner.Token,
		HeartBeat: runner.HeartBeat,
		Online:    string(runner.Online),
		Status:    string(runner.Status),
		EditedBy:  runner.EditedBy,
	}

	if err := db.Model(&Runner{}).Where("id = ? AND deleted = ?", runner.ID, 0).Updates(dbRunner).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	return nil
}

// GetByID gets runner by ID
func (s *Store) GetByID(ctx context.Context, id uint) (*runnerModel.Runner, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbRunner Runner
	if err := db.Where("id = ? AND deleted = ?", id, 0).First(&dbRunner).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbRunner), nil
}

// GetByToken gets runner by token
func (s *Store) GetByToken(ctx context.Context, token string) (*runnerModel.Runner, error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	var dbRunner Runner
	if err := db.Where("token = ? AND deleted = ?", token, 0).First(&dbRunner).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(&dbRunner), nil
}

// List returns a list of runners with optional filter
func (s *Store) List(ctx context.Context, filter *runnerModel.ReqFilter) (*database.PageResult[*runnerModel.Runner], error) {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	db = db.Model(&Runner{}).Where("deleted = ?", 0)

	if filter != nil {
		if filter.Query != "" {
			q := "%" + filter.Query + "%"
			db = db.Where("name LIKE ? OR hostname LIKE ? OR ip LIKE ?", q, q, q)
		}
		if filter.Online != "" {
			db = db.Where("online = ?", filter.Online)
		}
		if filter.Status != "" {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.OS != "" {
			db = db.Where("os = ?", filter.OS)
		}
		if filter.Arch != "" {
			db = db.Where("arch = ?", filter.Arch)
		}
	}

	// Apply sorting
	orderBy := "online ASC, status ASC, id DESC" // Default sort: enable, online first, then by id desc
	if filter != nil && filter.Sort != "" {
		sortDirection := "ASC"
		if strings.ToLower(filter.Order) == "desc" {
			sortDirection = "DESC"
		}
		if filter.Sort == "id" {
			orderBy = filter.Sort + " " + sortDirection
		} else {
			sortFields := strings.Split(filter.Sort, "-")
			orderBy = sortFields[0] + " " + sortDirection + ", " + sortFields[1] + " " + sortDirection + ", id DESC"
		}
	}
	db = db.Order(orderBy)

	var dbRunners []*Runner
	pageResult, err := database.PaginateDefault(ctx, db, filter.Page, filter.PageSize, &dbRunners)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// Convert database models to business models
	runners := make([]*runnerModel.Runner, len(dbRunners))
	for i, dbRunner := range dbRunners {
		runners[i] = mapToModel(dbRunner)
	}

	return &database.PageResult[*runnerModel.Runner]{
		Data:      runners,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

// Delete soft deletes a runner by ID
func (s *Store) Delete(ctx context.Context, id uint) error {
	db := dbtx.GetOrmAccessor(ctx, s.db)

	updateFields := map[string]any{
		"deleted": 1,
		"online":  "offline",
		"status":  "disable",
	}

	if err := db.Model(&Runner{}).Where("id = ?", id).Updates(updateFields).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}

	return nil
}

// mapToModel converts database model to business model
func mapToModel(r *Runner) *runnerModel.Runner {
	var labels []string
	if r.Labels != "" {
		labels = strings.Split(strings.Trim(r.Labels, ","), ",")
	}

	return &runnerModel.Runner{
		ID:          r.ID,
		Name:        r.Name,
		Desc:        r.Desc,
		Version:     r.Version,
		Hostname:    r.Hostname,
		IP:          r.IP,
		OS:          r.OS,
		Arch:        r.Arch,
		Labels:      labels,
		Runtime:     r.Runtime,
		Token:       r.Token,
		HeartBeat:   r.HeartBeat,
		Online:      constant.RunnerStatus(r.Online),
		Status:      constant.RunnerActiveStatus(r.Status),
		IsDefault:   r.IsDefault,
		CreatedBy:   r.CreatedBy,
		CreatedDate: r.CreatedDate,
		EditedBy:    r.EditedBy,
		EditedDate:  r.EditedDate,
	}
}
