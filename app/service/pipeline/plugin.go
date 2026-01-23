package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/resources"
)

var _ PluginService = (*Service)(nil)

type Service struct {
	pipelineStore pipeline.Interface
}

func NewPluginService(pipelineStore pipeline.Interface) *Service {
	return &Service{
		pipelineStore: pipelineStore,
	}
}

func (s *Service) Migrate(ctx context.Context, url string) error {
	log := logger.GetLogger(ctx)
	plugins, err := s.pipelineStore.GetAll(ctx)
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

	cnt := 0
	for _, file := range zipFile.File {
		exists, err := filepath.Match("**/plugins/*/*.yaml", file.Name)
		if err != nil {
			return fmt.Errorf("could not glob pattern: %w", err)
		}
		if !exists {
			continue
		}
		fc, err := file.Open()
		if err != nil {
			log.Errorf("could not open file %s: %v", file.Name, err)
			continue
		}
		defer fc.Close()
		var yaml bytes.Buffer
		_, err = io.Copy(&yaml, fc) //nolint:gosec // plugin source is configured via environment variables by user
		if err != nil {
			log.Errorf("could not read file contents %s: %v", file.Name, err)
			continue
		}
		// schema should be a valid config - if not log an error and continue.
		config, err := pipelineModel.ParseBytes(yaml.Bytes())
		if err != nil {
			log.Warnf("could not parse schema into valid config %s: %v", file.Name, err)
			continue
		}

		var desc string
		switch vv := config.Spec.(type) {
		case *pipelineModel.PluginStep:
			desc = vv.Description
		case *pipelineModel.PluginStage:
			desc = vv.Description
		default:
			log.Warnf("schema did not match a valid plugin schema %s: unexpected type %T", file.Name, config.Spec)
			continue
		}

		plugin := &pipelineModel.Plugin{
			Name: config.Name,
			Desc: desc,
			Type: config.Type,
			Yaml: yaml.String(),
		}

		// Try to read the logo if it exists in the same directory
		dir := filepath.Dir(file.Name)
		logoFile := filepath.Join(dir, "logo.svg")
		if lf, err := zipFile.Open(logoFile); err == nil { // if we can open the logo file
			var lbuf bytes.Buffer
			_, err = io.Copy(&lbuf, lf)
			if err != nil {
				log.Warnf("could not  copy logo file %s: %v", logoFile, err)
			} else {
				plugin.Logo = lbuf.String()
			}
		}

		// 如果数据库中已经存在插件,且插件规格未发生变化,则跳过upsert
		if p, ok := pluginMap[plugin.Name]; ok {
			if p.Matches(plugin) {
				continue
			}
		}

		if _, ok := pluginMap[plugin.Name]; ok {
			log.Infof("update plugin '%s'", plugin.Name)
			err = s.pipelineStore.Update(ctx, plugin)
			if err != nil {

				log.Warnf("could not update plugin %s: %v", plugin.Name, err)
				continue
			}
			log.Infof("detected changes: updated existing plugin entry %s", plugin.Name)
		} else {
			err = s.pipelineStore.Create(ctx, plugin)
			if err != nil {
				log.Warnf("could not create plugin in DB %s: %v", plugin.Name, err)
				continue
			}
			cnt++
		}
	}
	log.Infof("added %d new entries to plugins", cnt)

	return nil
}
