package artifact

import (
	"context"
	"errors"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ TreeNodeInterface = (*treeNode)(nil)

type treeNode struct{ db *gorm.DB }

func (c *treeNode) Create(ctx context.Context, newObj *model.CreateArtifactTreeNode) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	if !newObj.ParentID.Valid {
		row.ParentID = null.IntFrom(0)
	} else {
		row.ParentID = newObj.ParentID
	}
	row.ArtifactID = newObj.ArtifactID
	row.Name = newObj.Name
	row.Path = newObj.Path
	row.Type = string(newObj.Type)
	row.Format = string(newObj.Format)
	row.TargetID = newObj.TargetID
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	if err := c.db.WithContext(ctx).Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) GetByPath(ctx context.Context, artifactID int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error) {
	obj := &model.ArtifactTreeNode{ArtifactID: artifactID, Path: path, Format: format}
	if err := c.getNode(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *treeNode) GetByTarget(ctx context.Context, artifactID int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	q := artifactTreeNode{ArtifactID: artifactID, Type: string(nodeType), TargetID: id}
	if err := c.db.WithContext(ctx).Where(&q).Take(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) getNode(ctx context.Context, obj *model.ArtifactTreeNode) error {
	var row artifactTreeNode
	q := artifactTreeNode{ArtifactID: obj.ArtifactID, Path: obj.Path, Format: string(obj.Format)}
	if err := c.db.WithContext(ctx).Where(&q).Take(&row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	// 回写查询结果
	found := nodeToModel(&row)
	*obj = *found
	return nil
}

func (c *treeNode) ListFormats(ctx context.Context, artifactID int64) ([]*model.ArtifactTreeNode, error) {
	q := artifactTreeNode{ArtifactID: artifactID, Path: "/"}
	var rows []*artifactTreeNode
	if err := c.db.WithContext(ctx).Where(&q).Order("type").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodesToModel(rows), nil
}

func (c *treeNode) ListByParentId(ctx context.Context, parentID int64) ([]*model.ArtifactTreeNode, error) {
	if parentID == 0 {
		return nil, errors.New("parentID can not be zero")
	}
	var rows []*artifactTreeNode
	if err := c.db.WithContext(ctx).Where("parentID = ?", parentID).Order("nodeName").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodesToModel(rows), nil
}

func (c *treeNode) Update(ctx context.Context, upObj *model.ArtifactTreeNode) error {
	row := nodeFromModel(upObj)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *treeNode) DeleteById(ctx context.Context, nodeID int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactTreeNode{}, nodeID)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrPkgNoItemDeleted
	}
	return nil
}

func (c *treeNode) RecurseDeleteById(ctx context.Context, nodeID int64) error {
	subNodes, err := c.ListByParentId(ctx, nodeID)
	if err != nil {
		return err
	}
	for _, subNode := range subNodes {
		if err = c.RecurseDeleteById(ctx, subNode.ID); err != nil {
			return err
		}
	}
	return c.DeleteById(ctx, nodeID)
}
