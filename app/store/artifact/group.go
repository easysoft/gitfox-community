// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"context"
	"errors"

	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/database"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ GroupInterface = (*groups)(nil)

type groups struct{ db *gorm.DB }

const artifactGroupAncestorsQuery = `
WITH RECURSIVE art_group_ancestors(group_ancestor_id, group_ancestor_uid, group_ancestor_parent_id, group_ancestor_depth) AS (
	SELECT id, name, parentID, 0
	FROM ops_artifact_groups
	WHERE id = ?
	UNION
	SELECT g.id, g.name, g.parentID, a.group_ancestor_depth + 1
	FROM ops_artifact_groups g
	JOIN art_group_ancestors a ON g.id = a.group_ancestor_parent_id
)
`

// 从 owner 在指定 format 下的所有根组出发，递归展开未删除子树。
const artifactGroupOwnerDescendantsQuery = `
WITH RECURSIVE art_group_descendants(descendant_id,descendant_parentID,descendant_deleted) AS (
    SELECT id, parentID, deleted
    FROM ops_artifact_groups
    WHERE artifactLibID = ? AND type = ? AND parentID = 0
        AND (deleted IS NULL OR deleted = 0)
    UNION ALL
    SELECT id, parentID, deleted
    FROM ops_artifact_groups
    JOIN art_group_descendants ON parentID = descendant_id
    WHERE deleted IS NULL OR deleted = 0
)
`

// 从指定 group 开始，仅向下遍历未删除后代，不含其它平级分支。
const artifactGroupDescendantsQuery = `
WITH RECURSIVE art_group_descendants(descendant_id,descendant_parentID,descendant_deleted) AS (
    SELECT id, parentID, deleted
    FROM ops_artifact_groups
    WHERE id = ? AND (deleted IS NULL OR deleted = 0)
    UNION ALL
    SELECT id, parentID, deleted
    FROM ops_artifact_groups
    JOIN art_group_descendants ON parentID = descendant_id
    WHERE deleted IS NULL OR deleted = 0
)
`

func (c *groups) Create(ctx context.Context, newObj *model.CreateArtifactGroup) (*model.ArtifactGroup, error) {
	var row artifactGroup
	row.ParentID = newObj.ParentID
	row.ArtifactID = newObj.ArtifactID
	row.Name = newObj.Name
	row.Format = string(newObj.Format)
	row.EditedBy = newObj.EditedBy
	row.CreatedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) GetById(ctx context.Context, id int64) (*model.ArtifactGroup, error) {
	var row artifactGroup
	if err := c.db.WithContext(ctx).Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) GetRoot(ctx context.Context, artifactID int64, name string, format model.ArtifactFormat) (*model.ArtifactGroup, error) {
	var row artifactGroup
	q := artifactGroup{ArtifactID: artifactID, Name: name, Format: string(format)}
	if err := c.db.WithContext(ctx).Where(&q).Where("parentID = 0 and deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) GetSubNode(ctx context.Context, parentID int64, name string) (*model.ArtifactGroup, error) {
	var row artifactGroup
	q := artifactGroup{ParentID: null.IntFrom(parentID), Name: name}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").Take(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) GetEndpointByNames(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error) {
	obj := new(model.ArtifactGroupEndpoint)
	rootNode, err := c.GetRoot(ctx, artifactID, names[0], format)
	if err != nil {
		return nil, err
	}

	pathNum := len(names)
	if pathNum == 1 {
		obj.ArtifactGroup = rootNode
		return obj, nil
	}

	obj.Parents = append(obj.Parents, rootNode)
	for i := 1; i <= pathNum-1; i++ {
		node, e := c.GetSubNode(ctx, obj.Parents[i-1].ID, names[i])
		if e != nil {
			return nil, e
		}
		if i == pathNum-1 {
			obj.ArtifactGroup = node
		} else {
			obj.Parents = append(obj.Parents, node)
		}
	}
	return obj, nil
}

func (c *groups) getOrCreate(ctx context.Context, artifactID int64, name string, parentID null.Int, format model.ArtifactFormat, operator string) (*model.ArtifactGroup, error) {
	var dest *model.ArtifactGroup
	var err error

	if !parentID.Valid || parentID.Int64 == 0 {
		dest, err = c.GetRoot(ctx, artifactID, name, format)
	} else {
		dest, err = c.GetSubNode(ctx, parentID.Int64, name)
	}

	if err != nil {
		if errors.Is(err, pkgErrors.ErrResourceNotFound) {
			newObj := &model.CreateArtifactGroup{ArtifactID: artifactID, Name: name, Format: format, ParentID: parentID, CreatedBy: operator, EditedBy: operator}
			return c.Create(ctx, newObj)
		}
		return nil, err
	}
	return dest, nil
}

// RecursiveCreate 递归创建分组
func (c *groups) RecursiveCreate(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat, operator string) (*model.ArtifactGroupEndpoint, error) {
	obj := new(model.ArtifactGroupEndpoint)
	rootNode, err := c.getOrCreate(ctx, artifactID, names[0], null.IntFrom(0), format, operator)
	if err != nil {
		return nil, err
	}

	pathNum := len(names)
	if pathNum == 1 {
		obj.ArtifactGroup = rootNode
		return obj, nil
	}

	obj.Parents = append(obj.Parents, rootNode)
	for i := 1; i <= pathNum-1; i++ {
		node, e := c.getOrCreate(ctx, artifactID, names[i], null.IntFrom(obj.Parents[i-1].ID), format, operator)
		if e != nil {
			return nil, e
		}
		if i == pathNum-1 {
			obj.ArtifactGroup = node
		} else {
			obj.Parents = append(obj.Parents, node)
		}
	}
	return obj, nil
}

func (c *groups) GetEndpointById(ctx context.Context, id int64) (*model.ArtifactGroupEndpoint, error) {
	query := artifactGroupAncestorsQuery + `
		SELECT * ` + `
		FROM ops_artifact_groups INNER JOIN art_group_ancestors ON id = group_ancestor_id order by group_ancestor_depth desc`
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Raw(query, id).Scan(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	obj := new(model.ArtifactGroupEndpoint)
	num := len(rows)
	if num == 0 {
		return nil, database.ProcessSQLErrorf(ctx, gorm.ErrRecordNotFound)
	}
	dst := groupsToModel(rows)
	if num == 1 {
		obj.ArtifactGroup = dst[0]
	} else {
		obj.Parents = dst[0 : num-1]
		obj.ArtifactGroup = dst[num-1]
	}
	return obj, nil
}

func (c *groups) ListDescendantsForOwner(ctx context.Context, artifactID int64, format model.ArtifactFormat) ([]*model.ArtifactGroup, error) {
	query := artifactGroupOwnerDescendantsQuery + `
		SELECT ops_artifact_groups.* ` + `
		FROM ops_artifact_groups INNER JOIN art_group_descendants on id = descendant_id`
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Raw(query, artifactID, format).Scan(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

func (c *groups) ListDescendants(ctx context.Context, id int64) ([]*model.ArtifactGroup, error) {
	query := artifactGroupDescendantsQuery + `
		SELECT ops_artifact_groups.* ` + `
		FROM ops_artifact_groups INNER JOIN art_group_descendants gd ON id = descendant_id
		WHERE id != ?
		ORDER BY parentID ASC
		`
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Raw(query, id, id).Scan(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

func (c *groups) ListChildren(ctx context.Context, id int64) ([]*model.ArtifactGroup, error) {
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Where("parentID = ?", id).Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

// Search 按 SearchOption 组合条件查询 group 列表（与 packages.Search / versions.Search 对齐）。
// GC 链路依赖此方法按「软删 + editedDate < cutoff」筛选过期空壳 group。
func (c *groups) Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactGroup, error) {
	db := c.db.WithContext(ctx).Model(&artifactGroup{})
	for _, opt := range options {
		db = opt.Apply(db)
	}
	var rows []*artifactGroup
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

// FindRootByName 按 (artifactID, format, name) 查找根级 group（parentID=0 OR parentID IS NULL），
// 包含已软删记录；找不到时返回 (nil, nil)，便于调用方区分「不存在」与「系统错误」。
func (c *groups) FindRootByName(ctx context.Context, artifactID int64, name string, format model.ArtifactFormat) (*model.ArtifactGroup, error) {
	var row artifactGroup
	err := c.db.WithContext(ctx).
		Where("artifactLibID = ? AND name = ? AND type = ? and deleted = 0", artifactID, name, format).
		Where("parentID IS NULL OR parentID = 0").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) All(ctx context.Context) ([]*model.ArtifactGroup, error) {
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Where("deleted = 0").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

// SoftDeleteById 软删 group 并记录操作人到 editedBy（回收站列表 deletedBy 取此字段）。
func (c *groups) SoftDeleteById(ctx context.Context, groupID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	err := c.db.WithContext(ctx).Model(&artifactGroup{ID: groupID}).
		Updates(updates).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *groups) DeleteById(ctx context.Context, groupID int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactGroup{}, groupID)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgErrors.ErrResourceNotFound
	}
	return nil
}

// Update 更新 group 记录（全字段覆盖）。
func (c *groups) Update(ctx context.Context, upObj *model.ArtifactGroup) error {
	row := groupFromModel(upObj)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateNameByArtifactAndFormat 批量更新指定 view 下指定格式的 group name。
func (c *groups) UpdateNameByArtifactAndFormat(ctx context.Context, artifactID int64, oldName, newName string, format model.ArtifactFormat) error {
	err := c.db.WithContext(ctx).Model(&artifactGroup{}).
		Where("artifactLibID = ? AND name = ? AND type = ?", artifactID, oldName, string(format)).
		UpdateColumn("name", newName).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateName 更新分组名称。
func (c *groups) UpdateName(ctx context.Context, id int64, name string) error {
	if err := c.db.WithContext(ctx).Model(&artifactGroup{}).Where("id = ?", id).
		UpdateColumn("name", name).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// MoveLocation 调整 group 归属：写入新的 artifactLibID + parentID。
// RowsAffected==0 时返回 ErrResourceNotFound（group 已被删或 ID 不存在）。
func (c *groups) MoveLocation(ctx context.Context, groupID, artifactID, parentID int64) error {
	result := c.db.WithContext(ctx).Model(&artifactGroup{}).Where("id = ?", groupID).
		UpdateColumns(map[string]interface{}{
			"artifactLibID": artifactID,
			"parentID":      parentID,
		})
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgErrors.ErrResourceNotFound
	}
	return nil
}

// MoveArtifactLibByIDs 把一批 group 的 artifactLibID 改到目标 view，跨 view 移动子树时使用。
// ids 为空时直接返回 nil；不强制 RowsAffected 与 ids 数量匹配（已被软删的 group 会少更新）。
func (c *groups) MoveArtifactLibByIDs(ctx context.Context, groupIDs []int64, targetArtifactID int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	if err := c.db.WithContext(ctx).Model(&artifactGroup{}).
		Where("id IN ?", groupIDs).
		UpdateColumn("artifactLibID", targetArtifactID).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateNameByCondition 按条件批量更新分组名称（用于 container 制品库重命名时同步 group）。
func (c *groups) UpdateNameByCondition(ctx context.Context, artifactID int64, oldName, newName, format string) error {
	if err := c.db.WithContext(ctx).Model(&artifactGroup{}).
		Where("artifactLibID = ? AND name = ? AND type = ?", artifactID, oldName, format).
		UpdateColumn("name", newName).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
