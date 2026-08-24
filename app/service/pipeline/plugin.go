package pipeline

import (
	"context"

	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/file"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
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

	probe := data.NextFile(".yaml")
	data.CleanRead()

	cnt := 0
	if probe == nil {
		cnt = s.migrateNewFormat(ctx, data, groups, pluginMap, log)
	} else {
		cnt = s.migrateLegacyFormat(ctx, data, groups, pluginMap, log)
	}

	data.CleanRead()
	log.Infof("added %d new entries to plugins", cnt)

	return nil
}

// migrateLegacyFormat 迁移旧格式插件包：
// plugins/<group>/<plugin>/<name>.yaml（配置）+ <name>.json（UI）+ <name>.svg（Logo）。
// 插件名取文件名为 <name>。
func (s *Service) migrateLegacyFormat(ctx context.Context, data file.ZipData, groups []*pipelineModel.PluginGroup, pluginMap map[string]*pipelineModel.Plugin, log *logger.SugaredMateLogger) int {
	cnt := 0
	for {
		// NextFile 查找并返回 data 内的指定后缀的文件，返回后 id 记录在 Read 中，不重复返回，Read 中都已经返回则返回 nil
		yaml := data.NextFile(".yaml")
		if yaml == nil {
			break
		}

		// 跳过 macOS 系统生成的元数据文件
		if strings.Contains(yaml.Path, "__MACOSX/") || strings.Contains(yaml.Path, "/._") {
			continue
		}

		// 读取 yaml 文件
		yamlData, err := pipelineModel.ParseBytesSimple(yaml.Content.Bytes())
		if err != nil {
			log.Warnf("could not parse schema into valid config %s: %v", yaml.Path, err)
			continue
		}

		group := yaml.GetDirName(-2)

		// 检查插件是否属于某个分组
		groupID, err := GetPluginGroupID(groups, group)
		if err != nil {
			log.Warnf("could not get plugin group id %s: %v", group, err)
			continue
		}

		plugin := &pipelineModel.Plugin{
			GroupID: groupID,
			Name:    strings.ReplaceAll(yaml.Name, yaml.Suffix, ""),
			Desc:    yamlData.Spec.Description,
			Type:    yamlData.Type,
			Kind:    yamlData.Kind,
			Yaml:    yaml.Content.String(), // 直接保存整个 yaml 的 json 字符串
		}

		// 读取 json 文件保存 json 内容
		json := data.SearchPath(strings.ReplaceAll(yaml.Path, ".yaml", ".json"))
		if json == nil {
			log.Infof("could not find json file %s", strings.ReplaceAll(yaml.Path, ".yaml", ".json"))
			continue
		}
		plugin.Json = json.Content.String()

		// 读取 logo 文件保存 svg 内容
		svg := data.SearchPath(strings.ReplaceAll(json.Path, ".json", ".svg"))
		if svg != nil {
			plugin.Logo = svg.Content.String()
		}

		if s.upsertPlugin(ctx, plugin, pluginMap, log) {
			cnt++
		}
	}
	return cnt
}

// migrateNewFormat 迁移新格式插件包：
// plugins/<group>/<plugin>/bin.json（配置，等价旧 .yaml）+ ui.json（UI 表单，等价旧 .json）。
// 插件名取自 bin.json 内部 name 字段，而非文件名。
func (s *Service) migrateNewFormat(ctx context.Context, data file.ZipData, groups []*pipelineModel.PluginGroup, pluginMap map[string]*pipelineModel.Plugin, log *logger.SugaredMateLogger) int {
	cnt := 0
	for {
		// 新格式包内为 bin.json + ui.json，遍历 .json 文件仅处理 bin.json；
		// ui.json 由 SearchPath 按同目录配对获取
		bin := data.NextFile(".json")
		if bin == nil {
			break
		}

		// 仅处理 bin.json 配置文件，跳过 ui.json
		if bin.Name != "bin.json" {
			continue
		}

		// 跳过 macOS 系统生成的元数据文件
		if strings.Contains(bin.Path, "__MACOSX/") || strings.Contains(bin.Path, "/._") {
			continue
		}

		yamlData, err := pipelineModel.ParseBytesSimple(bin.Content.Bytes())
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

		// 新格式插件名取自 bin.json 内部 name 字段，文件名恒为 bin 不携带插件名
		name := yamlData.Name
		if name == "" {
			name = strings.ReplaceAll(bin.Name, bin.Suffix, "")
		}

		plugin := &pipelineModel.Plugin{
			GroupID: groupID,
			Name:    name,
			Desc:    yamlData.Spec.Description,
			Type:    yamlData.Type,
			Kind:    yamlData.Kind,
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
