package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// CreatePlugin 创建插件
func (s *Store) CreatePlugin(ctx context.Context, plugin *pipelineModel.Plugin) error {
	storePlugin := mapPluginToStore(plugin)
	if err := s.db.WithContext(ctx).Create(&storePlugin).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	plugin.ID = storePlugin.ID
	plugin.CreatedDate = storePlugin.CreatedDate
	plugin.EditedDate = storePlugin.EditedDate
	return nil
}

// UpdatePluginByName 更新插件（根据 name）
func (s *Store) UpdatePluginByName(ctx context.Context, plugin *pipelineModel.Plugin) error {
	storePlugin := mapPluginToStore(plugin)
	if err := s.db.WithContext(ctx).Model(&Plugin{}).Where("name = ? AND deleted = 0", plugin.Name).Updates(&storePlugin).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetPluginByName 根据插件名称获取插件
func (s *Store) GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error) {
	var plugin Plugin
	if err := s.db.WithContext(ctx).Where("name = ? AND deleted = 0", name).First(&plugin).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapPluginToModel(plugin), nil
}

// GetAllPlugin 获取所有插件
func (s *Store) GetAllPlugin(ctx context.Context) ([]*pipelineModel.Plugin, error) {
	var plugins []Plugin

	if err := s.db.WithContext(ctx).Where("deleted = 0").Find(&plugins).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*pipelineModel.Plugin, len(plugins))
	for i, p := range plugins {
		result[i] = mapPluginToModel(p)
	}
	return result, nil
}

// ListPlugin 查询插件列表（分页）
func (s *Store) ListPlugin(ctx context.Context, filter *pipelineModel.PluginFilter) (*database.PageResult[*pipelineModel.Plugin], error) {
	tx := s.db.WithContext(ctx).Model(&Plugin{}).Where("deleted = 0")

	// 类型过滤
	if filter.Type != "" {
		tx = tx.Where("type = ?", filter.Type)
	}

	// 执行分页查询
	var plugins []*Plugin
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &plugins)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 Model 层结构
	result := make([]*pipelineModel.Plugin, len(plugins))
	for i := range plugins {
		result[i] = mapPluginToModel(*plugins[i])
	}

	return &database.PageResult[*pipelineModel.Plugin]{
		Total:    pageResult.Total,
		Page:     pageResult.Page,
		PageSize: pageResult.PageSize,
		Data:     result,
	}, nil
}

// DeletePlugin 删除插件（硬删除）
func (s *Store) DeletePlugin(ctx context.Context, name string) error {
	if err := s.db.WithContext(ctx).Where("name = ?", name).Delete(&Plugin{}).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetAllPluginGroup 获取所有插件分组
func (s *Store) GetAllPluginGroup(ctx context.Context) ([]*pipelineModel.PluginGroup, error) {
	var groups []PluginGroup
	if err := s.db.WithContext(ctx).Where("deleted = 0").Find(&groups).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*pipelineModel.PluginGroup, len(groups))
	for i, g := range groups {
		result[i] = &pipelineModel.PluginGroup{
			ID:   g.ID,
			Name: g.Name,
			Desc: g.Desc,
		}
	}
	return result, nil
}

// PluginWithGroupRow 关联查询结果行
type PluginWithGroupRow struct {
	Plugin
	GroupName string `gorm:"column:group_name"`
	GroupDesc string `gorm:"column:group_desc"`
}

// GetAllWithGroup 获取所有插件及其分组信息（关联查询）
func (s *Store) GetAllWithGroup(ctx context.Context) ([]*pipelineModel.PluginWithGroup, error) {
	var rows []PluginWithGroupRow

	// 使用 JOIN 关联查询插件和分组
	err := s.db.WithContext(ctx).
		Table("ops_plugins p").
		Select("p.*, g.name as group_name, g.desc as group_desc").
		Joins("LEFT JOIN ops_plugin_group g ON p.groupID = g.id").
		Where("p.deleted = 0 AND g.deleted = 0").
		Find(&rows).Error

	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*pipelineModel.PluginWithGroup, len(rows))
	for i, row := range rows {
		result[i] = &pipelineModel.PluginWithGroup{
			Plugin:    *mapPluginToModel(row.Plugin),
			GroupName: row.GroupName,
			GroupDesc: row.GroupDesc,
		}
	}
	return result, nil
}

// mapPluginToModel 将 Store 层结构转换为 Model 层结构
func mapPluginToModel(p Plugin) *pipelineModel.Plugin {
	return &pipelineModel.Plugin{
		ID:          p.ID,
		GroupID:     p.GroupID,
		Name:        p.Name,
		Desc:        string(p.Desc),
		Logo:        string(p.Logo),
		Yaml:        string(p.Yaml),
		Json:        string(p.Json),
		Type:        pipelineModel.PluginType(p.Type),
		CreatedBy:   p.CreatedBy,
		CreatedDate: p.CreatedDate,
		EditedBy:    p.EditedBy,
		EditedDate:  p.EditedDate,
	}
}

// mapPluginToStore 将 Model 层结构转换为 Store 层结构
func mapPluginToStore(p *pipelineModel.Plugin) *Plugin {
	return &Plugin{
		ID:          p.ID,
		GroupID:     p.GroupID,
		Name:        p.Name,
		Desc:        []byte(p.Desc),
		Logo:        []byte(p.Logo),
		Yaml:        []byte(p.Yaml),
		Json:        []byte(p.Json),
		Type:        string(p.Type),
		Kind:        p.Kind,
		CreatedBy:   p.CreatedBy,
		CreatedDate: p.CreatedDate,
		EditedBy:    p.EditedBy,
		EditedDate:  p.EditedDate,
	}
}
