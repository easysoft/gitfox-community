package pipeline

import (
	"context"

	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/utils/file"
	"github.com/easysoft/gitfox-community/resources"
	"github.com/pkg/errors"
)

func (s *Service) Migrate(ctx context.Context, url string) error {
	log := logger.GetLogger(ctx)
	plugins, err := s.pipelineStore.GetAllPlugin(ctx)
	if err != nil {
		return err
	}

	pluginPkgPath, cancel, err := resources.DownloadCIPlugins(ctx, url)
	if err != nil {
		return err
	}
	defer cancel()

	pluginMap := map[string]*pipelineModel.Plugin{}
	for _, v := range plugins {
		pluginMap[v.Name] = v
	}

	data, err := file.GetArchiveData(pluginPkgPath)
	if err != nil {
		return err
	}

	groups, err := s.pipelineStore.GetAllPluginGroup(ctx)
	if err != nil {
		return err
	}

	cnt := s.migrateNewFormat(ctx, data, groups, pluginMap, log)
	log.Infof("added %d new entries to plugins", cnt)

	return nil
}

// migrateNewFormat 迁移新格式插件包：
// plugins/<group>/<plugin>/bin.json（配置）+ ui.json（UI 表单）。
// 文件名固定为 bin.json，插件名取自 bin.json 内部 name 字段。
func (s *Service) migrateNewFormat(ctx context.Context, data file.ZipData, groups []*pipelineModel.PluginGroup, pluginMap map[string]*pipelineModel.Plugin, log *logger.SugaredMateLogger) int {
	cnt := 0
	for {
		// 新格式包内文件名为固定的 bin.json + ui.json，按名取 bin.json，不重复返回
		bin := data.NextFileByName("bin.json")
		if bin == nil {
			break
		}

		// 跳过 macOS 系统生成的元数据文件
		if strings.Contains(bin.Path, "__MACOSX/") || strings.Contains(bin.Path, "/._") {
			continue
		}

		binData, err := pipelineModel.ParseBytesSimple(bin.Content.Bytes())
		if err != nil {
			log.Warnf("could not parse schema into valid config %s: %v", bin.Path, err)
			continue
		}

		group := bin.GetDirName(-2)

		// 检查插件是否属于某个分组
		groupID, err := GetPluginGroupID(groups, group)
		if err != nil {
			log.Warnf("could not get plugin group id %s: %v", group, err)
			continue
		}

		plugin := &pipelineModel.Plugin{
			GroupID: groupID,
			Name:    binData.Name,
			Desc:    binData.Spec.Description,
			Type:    binData.Type,
			Kind:    binData.Kind,
			Yaml:    bin.Content.String(),
		}

		// 读取同目录 ui.json 文件保存 UI 配置
		ui := data.SearchPath(strings.ReplaceAll(bin.Path, "bin.json", "ui.json"))
		if ui == nil {
			log.Infof("could not find ui.json file %s", strings.ReplaceAll(bin.Path, "bin.json", "ui.json"))
			continue
		}
		plugin.Json = ui.Content.String()

		if s.upsertPlugin(ctx, plugin, pluginMap, log) {
			cnt++
		}
	}
	return cnt
}

// upsertPlugin 对比插件是否发生变化，变化则更新，不存在则创建。
// 返回 true 表示本次为新增插件。
func (s *Service) upsertPlugin(ctx context.Context, plugin *pipelineModel.Plugin, pluginMap map[string]*pipelineModel.Plugin, log *logger.SugaredMateLogger) bool {
	if plugin.Name == "" {
		log.Warnf("could not get plugin name")
		return false
	}

	// 如果数据库中已经存在插件，且插件规格未发生变化，则跳过 upsert
	if p, ok := pluginMap[plugin.Name]; ok {
		if p.Matches(plugin) {
			return false
		}
	}

	// 落库
	if _, ok := pluginMap[plugin.Name]; ok {
		log.Infof("update plugin '%s'", plugin.Name)
		err := s.pipelineStore.UpdatePluginByName(ctx, plugin)
		if err != nil {
			log.Warnf("could not update plugin %s: %v", plugin.Name, err)
			return false
		}
		log.Infof("detected changes: updated existing plugin entry %s", plugin.Name)
		return false
	}

	err := s.pipelineStore.CreatePlugin(ctx, plugin)
	if err != nil {
		log.Warnf("could not create plugin in DB %s: %v", plugin.Name, err)
		return false
	}
	return true
}

// Resolve resolves the pipeline plugin configuration.
// kind 插件种类，typ 插件类型，pluginName 插件名称
func (s *Service) Resolve(ctx context.Context, kind, typ, pluginName string) (*pipelineModel.Yaml, error) {

	k, err := constant.ParseResolverKind(kind)
	if err != nil {
		return nil, err
	}

	t, err := constant.ParseResolverType(typ)
	if err != nil {
		return nil, err
	}

	if k == constant.ResolverKindPlugin && t != constant.ResolverTypeStep {
		return nil, errors.Errorf("only step level plugins are currently supported")
	}

	if k == constant.ResolverKindPlugin {
		plugin, err := s.pipelineStore.GetPluginByName(ctx, pluginName)
		if err != nil {
			return nil, errors.Wrapf(err, "could not lookup plugin")
		}

		// Convert plugin to v1yaml spec
		config, err := pipelineModel.ParseString(plugin.Yaml)
		if err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal plugin to v1yaml spec")
		}
		return config, nil
	}

	return nil, errors.Errorf("could not resolve plugin %s", pluginName)
}

// GetPluginByName 根据插件名称获取插件
func (s *Service) GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error) {
	return s.pipelineStore.GetPluginByName(ctx, name)
}
