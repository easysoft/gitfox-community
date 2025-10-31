package space

import (
	"context"
	"time"

	modelSpace "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

type SpaceStore struct {
	db *gorm.DB
}

func NewSpaceStore(db *gorm.DB) *SpaceStore {
	return &SpaceStore{db: db}
}

func (r *SpaceStore) Create(ctx context.Context, req modelSpace.CreateSpaceRequest) (*modelSpace.SpaceResponse, error) {

	now := time.Now()
	data := Space{
		Name:        req.Name,
		Owner:       req.Owner,
		Permission:  req.Permission,
		Desc:        req.Desc,
		CreatedBy:   req.CreatedBy,
		CreatedDate: &now,
		EditedBy:    req.CreatedBy,
		EditedDate:  &now,
	}

	if err := r.db.WithContext(ctx).Create(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(data), nil
}

func (r *SpaceStore) Update(ctx context.Context, id uint16, req modelSpace.UpdateSpaceRequest) (*modelSpace.SpaceResponse, error) {
	var out Space

	now := time.Now()
	updateData := make(map[string]any)

	// EditedBy 和 EditedDate 是必填字段
	updateData["editedBy"] = req.EditedBy
	updateData["editedDate"] = &now

	// 只更新非空字段
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Owner != "" {
		updateData["owner"] = req.Owner
	}
	if req.Permission != "" {
		updateData["acl"] = req.Permission
	}
	if req.Desc != "" {
		updateData["desc"] = req.Desc
	}

	// 使用事务确保查询和更新的原子性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询出要更新的记录
		if err := tx.Where("id = ?", id).First(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 提取 map 中的所有字段名,用于 Select
		fields := make([]string, 0, len(updateData))
		for field := range updateData {
			fields = append(fields, field)
		}

		// 使用 Select 明确指定要更新的字段,支持零值更新
		// Updates 会直接更新 out 对象（传入指针）
		if err := tx.Model(&out).
			Select(fields).
			Updates(updateData).Error; err != nil {
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

// mapToModel 将 Space 转换为 SpaceResponse
func mapToModel(data Space) *modelSpace.SpaceResponse {
	return &modelSpace.SpaceResponse{
		ID:          data.ID,
		Name:        data.Name,
		Owner:       data.Owner,
		Permission:  data.Permission,
		Desc:        data.Desc,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
	}
}
