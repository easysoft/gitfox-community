package artifact

import (
	"context"
	"errors"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ GroupInterface = (*groups)(nil)

type groups struct{ db *gorm.DB }

const artifactGroupAncestorsQuery = `
WITH RECURSIVE art_group_ancestors(group_ancestor_id, group_ancestor_uid, group_ancestor_parent_id) AS (
	SELECT id, name, parentID
	FROM ops_artifact_groups
	WHERE id = ?
	UNION
	SELECT id, name, parentID
	FROM ops_artifact_groups
	JOIN art_group_ancestors ON id = group_ancestor_parent_id
)
`

// 从 owner 在指定 format 下的所有根组出发，递归展开未删除子树。
const artifactGroupOwnerDescendantsQuery = `
WITH RECURSIVE art_group_descendants(descendant_id,descendant_parentID,descendant_deleted) AS (
    SELECT id, parentID, deleted
    FROM ops_artifact_groups
    WHERE artifactID = ? AND format = ? AND parentID = 0
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
	if err := c.db.WithContext(ctx).Where(&q).Where("parentID = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupToModel(&row), nil
}

func (c *groups) GetSubNode(ctx context.Context, parentId int64, name string) (*model.ArtifactGroup, error) {
	var row artifactGroup
	q := artifactGroup{ParentID: null.IntFrom(parentId), Name: name}
	if err := c.db.WithContext(ctx).Where(&q).Take(&row).Error; err != nil {
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

func (c *groups) getOrCreate(ctx context.Context, artifactID int64, name string, parentId null.Int, format model.ArtifactFormat) (*model.ArtifactGroup, error) {
	var dest *model.ArtifactGroup
	var err error

	if !parentId.Valid || parentId.Int64 == 0 {
		dest, err = c.GetRoot(ctx, artifactID, name, format)
	} else {
		dest, err = c.GetSubNode(ctx, parentId.Int64, name)
	}

	if err != nil {
		if errors.Is(err, pkgErrors.ErrResourceNotFound) {
			newObj := &model.CreateArtifactGroup{ArtifactID: artifactID, Name: name, Format: format, ParentID: parentId}
			return c.Create(ctx, newObj)
		}
		return nil, err
	}
	return dest, nil
}

func (c *groups) RecursiveCreate(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error) {
	obj := new(model.ArtifactGroupEndpoint)
	rootNode, err := c.getOrCreate(ctx, artifactID, names[0], null.IntFrom(0), format)
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
		node, e := c.getOrCreate(ctx, artifactID, names[i], null.IntFrom(obj.Parents[i-1].ID), format)
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
		FROM ops_artifact_groups INNER JOIN art_group_ancestors ON id = group_ancestor_id order by id`
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

func (c *groups) All(ctx context.Context) ([]*model.ArtifactGroup, error) {
	var rows []*artifactGroup
	if err := c.db.WithContext(ctx).Where("deleted = 0").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return groupsToModel(rows), nil
}

func (c *groups) SoftDeleteById(ctx context.Context, groupId int64) error {
	err := c.db.WithContext(ctx).Model(&artifactGroup{ID: groupId}).
		UpdateColumn("deleted", 1).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// DeleteByArtifactID 硬删指定 view 下的所有 group 记录（含已软删的）。
func (c *groups) DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error) {
	result := c.db.WithContext(ctx).Where("artifactID = ?", artifactID).Delete(&artifactGroup{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
