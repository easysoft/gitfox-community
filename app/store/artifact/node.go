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

var _ TreeNodeInterface = (*treeNode)(nil)

type treeNode struct{ db *gorm.DB }

func (c *treeNode) Create(ctx context.Context, newObj *model.CreateArtifactTreeNode) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	row.ParentID = newObj.ParentID
	row.ViewID = newObj.ViewID
	row.Name = newObj.Name
	row.Path = newObj.Path
	row.Type = string(newObj.Type)
	row.Format = string(newObj.Format)
	row.TargetId = newObj.TargetId
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) RecurseCreate(ctx context.Context, obj *model.ArtifactTreeNode) error {
	if err := c.getNode(ctx, obj); err == nil {
		return nil
	} else if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return err
	}

	if obj.IsRoot() {
		created, err := c.Create(ctx, &model.CreateArtifactTreeNode{
			ParentID: obj.ParentID, ViewID: obj.ViewID, Name: obj.Name,
			Path: obj.Path, Type: obj.Type, Format: obj.Format, TargetId: obj.TargetId,
		})
		if err != nil {
			return err
		}
		obj.ID = created.ID
		return nil
	}

	p := obj.Parent()
	if err := c.RecurseCreate(ctx, p); err != nil {
		return err
	}
	obj.ParentID = null.IntFrom(p.ID)
	created, err := c.Create(ctx, &model.CreateArtifactTreeNode{
		ParentID: obj.ParentID, ViewID: obj.ViewID, Name: obj.Name,
		Path: obj.Path, Type: obj.Type, Format: obj.Format, TargetId: obj.TargetId,
	})
	if err != nil {
		return err
	}
	obj.ID = created.ID
	return nil
}

func (c *treeNode) GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	if err := c.db.WithContext(ctx).Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) GetByPath(ctx context.Context, viewId int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error) {
	obj := &model.ArtifactTreeNode{ViewID: viewId, Path: path, Format: format}
	if err := c.getNode(ctx, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *treeNode) GetByTarget(ctx context.Context, viewId int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error) {
	var row artifactTreeNode
	q := artifactTreeNode{ViewID: viewId, Type: string(nodeType), TargetId: null.IntFrom(id)}
	if err := c.db.WithContext(ctx).Where(&q).Take(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodeToModel(&row), nil
}

func (c *treeNode) getNode(ctx context.Context, obj *model.ArtifactTreeNode) error {
	var row artifactTreeNode
	q := artifactTreeNode{ViewID: obj.ViewID, Path: obj.Path, Format: string(obj.Format)}
	if err := c.db.WithContext(ctx).Where(&q).Take(&row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	// 回写查询结果
	found := nodeToModel(&row)
	*obj = *found
	return nil
}

func (c *treeNode) ListFormats(ctx context.Context, viewId int64) ([]*model.ArtifactTreeNode, error) {
	q := artifactTreeNode{ViewID: viewId, Path: "/"}
	var rows []*artifactTreeNode
	if err := c.db.WithContext(ctx).Where(&q).Order("format").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return nodesToModel(rows), nil
}

func (c *treeNode) ListByParentId(ctx context.Context, parentId int64) ([]*model.ArtifactTreeNode, error) {
	if parentId == 0 {
		return nil, errors.New("parentId can not be zero")
	}
	var rows []*artifactTreeNode
	if err := c.db.WithContext(ctx).Where("parentID = ?", parentId).Order("name").Find(&rows).Error; err != nil {
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

func (c *treeNode) DeleteById(ctx context.Context, nodeId int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactTreeNode{}, nodeId)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrPkgNoItemDeleted
	}
	return nil
}

func (c *treeNode) RecurseDeleteById(ctx context.Context, nodeId int64) error {
	subNodes, err := c.ListByParentId(ctx, nodeId)
	if err != nil {
		return err
	}
	for _, subNode := range subNodes {
		if err = c.RecurseDeleteById(ctx, subNode.ID); err != nil {
			return err
		}
	}
	return c.DeleteById(ctx, nodeId)
}

// DeleteByViewID 硬删指定 view 下的所有树节点记录。
func (c *treeNode) DeleteByViewID(ctx context.Context, viewId int64) (int64, error) {
	result := c.db.WithContext(ctx).Where("viewID = ?", viewId).Delete(&artifactTreeNode{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
