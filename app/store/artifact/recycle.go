// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"

	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/database"
)

// =============================================================================
// 回收站列表查询：5 个表（view / group / package / version / asset）各一个独立方法。
// 入参均为 (page, pageSize)，按 editedDate desc 返回；deleted=1 即「在回收站」，
// 不再做时间窗过滤——前端按 expiresAt 自行展示距离 GC 多久。
// =============================================================================

// recycleApply 通用：deleted=1 + 按 editedDate desc + 分页，返回 (offset,limit) 计算结果。
func recyclePaging(page, pageSize int) (offset, limit int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return (page - 1) * pageSize, pageSize
}

// ListRecycle 列出回收站中的制品库（ops_artifact 软删记录）。
func (c *libs) ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactView, int64, error) {
	offset, limit := recyclePaging(page, pageSize)
	var (
		rows  []*artifactLib
		total int64
	)
	stmt := c.db.WithContext(ctx).Model(&artifactLib{}).Where("deleted = 1")
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	if err := stmt.Order("editedDate desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	out := make([]*model.ArtifactView, 0, len(rows))
	for _, r := range rows {
		out = append(out, viewToModel(r))
	}
	return out, total, nil
}

// ListRecycle 列出回收站中的分组。
func (c *groups) ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactGroup, int64, error) {
	offset, limit := recyclePaging(page, pageSize)
	var (
		rows  []*artifactGroup
		total int64
	)
	stmt := c.db.WithContext(ctx).Model(&artifactGroup{}).Where("deleted = 1")
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	if err := stmt.Order("editedDate desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), total, nil
}

// ListRecycle 列出回收站中的制品包。
func (c *packages) ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactPackage, int64, error) {
	offset, limit := recyclePaging(page, pageSize)
	var (
		rows  []*artifactPackage
		total int64
	)
	stmt := c.db.WithContext(ctx).Model(&artifactPackage{}).Where("deleted = 1")
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	if err := stmt.Order("editedDate desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	out := make([]*model.ArtifactPackage, 0, len(rows))
	for _, r := range rows {
		out = append(out, packageToModel(r))
	}
	return out, total, nil
}

// ListRecycle 列出回收站中的制品版本。
func (c *versions) ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactVersion, int64, error) {
	offset, limit := recyclePaging(page, pageSize)
	var (
		rows  []*artifactVersion
		total int64
	)
	stmt := c.db.WithContext(ctx).Model(&artifactVersion{}).Where("deleted = 1")
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	if err := stmt.Order("editedDate desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	out := make([]*model.ArtifactVersion, 0, len(rows))
	for _, r := range rows {
		out = append(out, versionToModel(r))
	}
	return out, total, nil
}

// ListRecycle 列出回收站中的制品资产。
func (c *assets) ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactAsset, int64, error) {
	offset, limit := recyclePaging(page, pageSize)
	var (
		rows  []*ArtifactAsset
		total int64
	)
	stmt := c.db.WithContext(ctx).Model(&ArtifactAsset{}).Where("deleted = 1")
	if err := stmt.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	if err := stmt.Order("editedDate desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}
	out := make([]*model.ArtifactAsset, 0, len(rows))
	for _, r := range rows {
		out = append(out, assetToModel(r))
	}
	return out, total, nil
}
