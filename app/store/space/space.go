// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package space

import (
	"context"
	"strings"

	spaceModel "github.com/easysoft/gitfox-community/app/model/space"

	"github.com/easysoft/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

type SpaceStore struct {
	db *gorm.DB
}

func NewSpaceStore(db *gorm.DB) *SpaceStore {
	return &SpaceStore{db: db}
}

func (s *SpaceStore) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error) {

	data := Space{
		Name:      req.Name,
		Code:      req.Code,
		Auth:      req.Auth,
		ACL:       req.ACL,
		Desc:      &req.Desc,
		CreatedBy: req.CreatedBy,
		EditedBy:  req.CreatedBy,
	}

	if err := s.db.WithContext(ctx).Create(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(data), nil
}

func (s *SpaceStore) Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	var out Space
	out.ID = id
	// 使用事务确保查询和更新的原子性
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询出要更新的记录
		if err := tx.First(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		// EditedBy
		out.EditedBy = req.EditedBy

		if req.Name != "" {
			out.Name = req.Name
		}
		if req.Auth != "" {
			out.Auth = req.Auth
		}
		if req.ACL != "" {
			out.ACL = req.ACL
		}
		out.Desc = &req.Desc

		// 3. 更新（BeforeUpdate钩子会自动填充EditedDate，无需手动处理）
		if err := tx.Updates(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 将 Space 转换为 SpaceResponse
	return mapToModel(out), nil
}

func (s *SpaceStore) List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*Space], error) {
	tx := s.db.WithContext(ctx).Model(&Space{})
	if filter.Code != "" {
		tx = tx.Where("code LIKE ?", "%"+filter.Code+"%")
	}
	if filter.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.ACL != "" {
		tx = tx.Where("acl = ?", filter.ACL)
	}
	if !filter.Deleted {
		tx = tx.Where("deleted = ?", 0)
	}

	ids := strings.Split(filter.IDs, ",")
	// 查询包含所有公开空间
	if filter.IsOpen {
		// 嵌套Where：强制 (acl='open' OR id IN (ids)) 作为整体，与前面的条件做AND
		if len(filter.IDs) > 0 {
			tx = tx.Where("acl = ? OR id IN ?", "open", ids)
		} else {
			tx = tx.Where("acl = ?", "open")
		}
	} else if len(filter.IDs) > 0 {
		tx = tx.Where("id IN ?", ids)
	}

	tx = tx.Distinct().Order("CASE acl WHEN 'private' THEN 0 WHEN 'open' THEN 1 ELSE 2 END, id DESC")

	// 2. 调用通用分页工具（NeedTotal默认true，自动统计总条数）
	var spaces []*Space
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &spaces)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return pageResult, nil
}

// GetByID 根据ID查询单个空间
func (r *SpaceStore) GetByID(ctx context.Context, id uint) (*Space, error) {
	var out Space
	out.ID = id

	if err := r.db.WithContext(ctx).Where("deleted = ?", 0).First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return &out, nil
}

// GetByCode 根据标识码查询单个空间。
func (r *SpaceStore) GetByCode(ctx context.Context, code string) (*Space, error) {
	var out Space
	if err := r.db.WithContext(ctx).Where("code = ? AND deleted = ?", code, 0).First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &out, nil
}

// GetByName 根据名称查询单个空间。
func (r *SpaceStore) GetByName(ctx context.Context, name string) (*Space, error) {
	var out Space
	if err := r.db.WithContext(ctx).Where("name = ? AND deleted = ?", name, 0).First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &out, nil
}

// HasRepositories 检查空间是否有关联的代码库
func (r *SpaceStore) HasRepositories(ctx context.Context, spaceID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("ops_repo").
		Where("spaceID = ? AND deleted = 0", spaceID).
		Count(&count).Error; err != nil {
		return false, database.ProcessSQLErrorf(ctx, err)
	}
	return count > 0, nil
}

// HasArtifacts 检查空间是否有关联的制品库
func (r *SpaceStore) HasArtifacts(ctx context.Context, spaceID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("ops_artifact_libs").
		Where("spaceID = ? AND deleted = 0", spaceID).
		Count(&count).Error; err != nil {
		return false, database.ProcessSQLErrorf(ctx, err)
	}
	return count > 0, nil
}

// Delete 逻辑删除空间
func (r *SpaceStore) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 查询是否删除.
		var space Space
		if err := tx.Where("id = ?", id).Where("deleted = ?", 0).First(&space).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 删除操作
		space.Deleted = 1
		if err := tx.Model(&space).Select("deleted", "editedDate").Updates(&space).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// Restore 批量恢复已删除的空间
func (r *SpaceStore) Restore(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 状态校验
		var space Space
		if err := tx.Where("id = ?", id).Where("deleted = ?", 1).First(&space).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		space.Deleted = 0

		if err := tx.Model(&space).Select("deleted", "editedDate").Updates(&space).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})

}

// mapToModel 将 Space 转换为 SpaceResponse
func mapToModel(data Space) *spaceModel.SpaceResponse {
	desc := ""
	if data.Desc != nil {
		desc = *data.Desc
	}
	return &spaceModel.SpaceResponse{
		ID:          data.ID,
		Name:        data.Name,
		Code:        data.Code,
		Auth:        data.Auth,
		ACL:         data.ACL,
		Desc:        desc,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
	}
}
