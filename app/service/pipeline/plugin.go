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

	zipFile, cancel, err := resources.LoadAllCIPlugins(ctx, url)
	if err != nil {
		return err
	}
	defer cancel()

	pluginMap := map[string]*pipelineModel.Plugin{}
	for _, v := range plugins {
		pluginMap[v.Name] = v
	}

	data, err := file.GetZipData(zipFile)
	if err != nil {
		return err
	}
	_ = data
	cnt := 0

	groups, err := s.pipelineStore.GetAllPluginGroup(ctx)
	if err != nil {
		return err
	}

	for {
		// NextFile 查找并返回data 内的指定后缀的文件，返回后id记录在 Read 中，不重复返回 ，Read 中都已经返回则返回nil
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
			Yaml:    yaml.Content.String(), // 直接保存整个yaml 的json 字符串
		}

		// 读取 json 文件 保存json 内容
		json := data.SearchPath(strings.ReplaceAll(yaml.Path, ".yaml", ".json"))
		if json == nil {
			log.Infof("could not find json file %s", strings.ReplaceAll(yaml.Path, ".yaml", ".json"))
			continue
		}
		plugin.Json = json.Content.String()

		// 读取 logo 文件 保存svg 内容
		svg := data.SearchPath(strings.ReplaceAll(json.Path, ".json", ".svg"))
		if svg != nil {
			plugin.Logo = svg.Content.String()
		}

		// 如果数据库中已经存在插件,且插件规格未发生变化,则跳过upsert
		if p, ok := pluginMap[plugin.Name]; ok {
			if p.Matches(plugin) {
				continue
			}
		}

		if name := plugin.Name; name == "" {
			log.Warnf("could not get plugin name %s", yaml.Path)
			continue
		}

		// 落库
		if _, ok := pluginMap[plugin.Name]; ok {
			log.Infof("update plugin '%s'", plugin.Name)
			err = s.pipelineStore.UpdatePluginByName(ctx, plugin)
			if err != nil {

				log.Warnf("could not update plugin %s: %v", plugin.Name, err)
				continue
			}
			log.Infof("detected changes: updated existing plugin entry %s", plugin.Name)
		} else {
			err = s.pipelineStore.CreatePlugin(ctx, plugin)
			if err != nil {
				log.Warnf("could not create plugin in DB %s: %v", plugin.Name, err)
				continue
			}
			cnt++
		}
	}

	data.CleanRead()
	log.Infof("added %d new entries to plugins", cnt)

	return nil
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
