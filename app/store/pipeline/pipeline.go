// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package pipeline

import (
	"context"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
)

// CreatePipeline 创建流水线
func (s *Store) CreatePipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) (*pipelineModel.Pipeline, error) {
	dbPipeline := mapModelToPipeline(pipeline)
	if err := s.db.WithContext(ctx).Create(dbPipeline).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineToModel(dbPipeline), nil
}

// UpdatePipeline 更新流水线
func (s *Store) UpdatePipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) error {
	dbPipeline := mapModelToPipeline(pipeline)
	if err := s.db.WithContext(ctx).Model(&Pipeline{}).Where("id = ?", pipeline.ID).Updates(dbPipeline).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetPipelineByID 根据流水线ID获取流水线（deleted=0）
func (s *Store) GetPipelineByID(ctx context.Context, id int64) (*pipelineModel.Pipeline, error) {
	var pipeline Pipeline
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&pipeline).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineToModel(&pipeline), nil
}

// GetPipelineByRepoIDAndName 根据仓库ID和流水线名称获取流水线（deleted=0）
func (s *Store) GetPipelineByRepoIDAndName(ctx context.Context, repoID int64, name string) (*pipelineModel.Pipeline, error) {
	var pipeline Pipeline
	if err := s.db.WithContext(ctx).Where("repoID = ? AND name = ? AND deleted = 0", repoID, name).First(&pipeline).Error; err != nil {
		sqlErr := database.ProcessSQLErrorf(ctx, err)
		if errors.Is(sqlErr, errors.ErrResourceNotFound) {
			return nil, nil
		}
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPipelineToModel(&pipeline), nil
}

// mapModelToPipeline 将流水线模型映射为数据库模型
func mapModelToPipeline(p *pipelineModel.Pipeline) *Pipeline {
	return &Pipeline{
		ID:               p.ID,
		Name:             p.Name,
		Engine:           p.Engine,
		ProviderID:       p.ProviderID,
		Scope:            p.Scope,
		SpaceID:          p.SpaceID,
		RepoID:           p.RepoID,
		Desc:             p.Desc,
		Status:           p.Status,
		LatestVersion:    p.LatestVersion,
		DefaultBranch:    p.DefaultBranch,
		YamlPath:         p.YamlPath,
		CustomParam:      p.CustomParam,
		ExternalPipeline: p.ExternalPipeline,
		LastExec:         p.LastExec,
		LastResult:       p.LastResult,
		CreatedBy:        p.CreatedBy,
		CreatedDate:      p.CreatedDate,
		EditedBy:         p.EditedBy,
		EditedDate:       p.EditedDate,
		Deleted:          uint8(p.Deleted),
	}
}

// mapPipelineToModel 将数据库模型映射为流水线模型
func mapPipelineToModel(p *Pipeline) *pipelineModel.Pipeline {
	return &pipelineModel.Pipeline{
		ID:               p.ID,
		Name:             p.Name,
		Engine:           p.Engine,
		ProviderID:       p.ProviderID,
		Scope:            p.Scope,
		SpaceID:          p.SpaceID,
		RepoID:           p.RepoID,
		Desc:             p.Desc,
		Status:           p.Status,
		LatestVersion:    p.LatestVersion,
		DefaultBranch:    p.DefaultBranch,
		YamlPath:         p.YamlPath,
		CustomParam:      p.CustomParam,
		ExternalPipeline: p.ExternalPipeline,
		LastExec:         p.LastExec,
		LastResult:       p.LastResult,
		CreatedBy:        p.CreatedBy,
		CreatedDate:      p.CreatedDate,
		EditedBy:         p.EditedBy,
		EditedDate:       p.EditedDate,
		Deleted:          int8(p.Deleted),
	}
}
