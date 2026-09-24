// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"context"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
)

// GetGroupedPlugins 获取按分组归类的插件列表
func (s *Service) GetGroupedPlugins(ctx context.Context) ([]*pipelineModel.GroupedPlugins, error) {
	// 一次性查询所有插件及其分组信息
	pluginsWithGroup, err := s.pipelineStore.GetAllWithGroup(ctx)
	if err != nil {
		return nil, err
	}

	// 按分组ID进行分组，使用 map 保持分组顺序
	groupMap := make(map[int64]*pipelineModel.GroupedPlugins)
	groupOrder := make([]int64, 0)

	for _, p := range pluginsWithGroup {
		// 过滤掉 default 分组
		if p.GroupName == pipelineModel.DefaultPluginGroup {
			continue
		}

		// 如果分组不存在，创建新分组
		if _, exists := groupMap[p.GroupID]; !exists {
			groupMap[p.GroupID] = &pipelineModel.GroupedPlugins{
				GroupID:   p.GroupID,
				GroupName: p.GroupName,
				Desc:      p.GroupDesc,
				Plugins:   make([]*pipelineModel.SimpleStep, 0),
			}
			groupOrder = append(groupOrder, p.GroupID)
		}

		// 添加插件到对应分组
		groupMap[p.GroupID].Plugins = append(groupMap[p.GroupID].Plugins, &pipelineModel.SimpleStep{
			Name:   p.Name,
			NameCn: p.Desc,
			Svg:    p.Logo,
			Kind:   p.Kind,
		})
	}

	// 按原始顺序构建结果
	result := make([]*pipelineModel.GroupedPlugins, 0, len(groupOrder))
	for _, groupID := range groupOrder {
		result = append(result, groupMap[groupID])
	}

	return result, nil
}

// GetPluginGroupID 根据分组名称匹配分组ID
func GetPluginGroupID(groups []*pipelineModel.PluginGroup, groupName string) (int64, error) {

	for _, v := range groups {
		if v.Name == groupName {
			return v.ID, nil
		}
	}

	for _, v := range groups {
		if v.Name == pipelineModel.DefaultPluginGroup {
			return v.ID, nil
		}
	}

	return 0, nil
}
