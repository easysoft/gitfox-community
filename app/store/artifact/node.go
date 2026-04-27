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

func (c *treeNode) Create(ctx context.Context, newObj *model.ArtifactTreeNode) error {
	if err := c.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *treeNode) RecurseCreate(ctx context.Context, obj *model.ArtifactTreeNode) error {
	if err := c.getNode(ctx, obj); err == nil {
		return nil
	} else if !errors.Is(err, pkgErrors.ErrResourceNotFound) {
		return err
	}

	if obj.IsRoot() {
		return c.Create(ctx, obj)
	}

	p := obj.Parent()
	if err := c.RecurseCreate(ctx, p); err != nil {
		return err
	}
	obj.ParentID = null.IntFrom(p.ID)
	return c.Create(ctx, obj)
}

func (c *treeNode) GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error) {
	var obj model.ArtifactTreeNode
	if err := c.db.WithContext(ctx).Take(&obj, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &obj, nil
}

func (c *treeNode) GetByPath(ctx context.Context, viewId int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error) {
	result := model.ArtifactTreeNode{ViewID: viewId, Path: path, Format: format}
	if err := c.getNode(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *treeNode) GetByTarget(ctx context.Context, viewId int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error) {
	var dst model.ArtifactTreeNode
	q := model.ArtifactTreeNode{ViewID: viewId, Type: nodeType, TargetId: null.IntFrom(id)}
	if err := c.db.WithContext(ctx).Where(&q).Take(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &dst, nil
}

func (c *treeNode) getNode(ctx context.Context, obj *model.ArtifactTreeNode) error {
	q := model.ArtifactTreeNode{ViewID: obj.ViewID, Path: obj.Path, Format: obj.Format}
	if err := c.db.WithContext(ctx).Where(&q).Take(obj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *treeNode) ListFormats(ctx context.Context, viewId int64) ([]*model.ArtifactTreeNode, error) {
	q := model.ArtifactTreeNode{ViewID: viewId, Path: "/"}
	var data []*model.ArtifactTreeNode
	if err := c.db.WithContext(ctx).Where(&q).Order("format").Find(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return data, nil
}

func (c *treeNode) ListByParentId(ctx context.Context, parentId int64) ([]*model.ArtifactTreeNode, error) {
	if parentId == 0 {
		return nil, errors.New("parentId can not be zero")
	}
	var data []*model.ArtifactTreeNode
	if err := c.db.WithContext(ctx).Where("parentID = ?", parentId).Order("name").Find(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return data, nil
}

func (c *treeNode) Update(ctx context.Context, upObj *model.ArtifactTreeNode) error {
	if err := c.db.WithContext(ctx).Save(upObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *treeNode) DeleteById(ctx context.Context, nodeId int64) error {
	result := c.db.WithContext(ctx).Delete(&model.ArtifactTreeNode{}, nodeId)
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
	result := c.db.WithContext(ctx).Where("viewID = ?", viewId).Delete(&model.ArtifactTreeNode{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
