package manager

import (
	"context"
	"encoding/json"
	"fmt"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	v1yaml "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec/parse/walk"

	"github.com/ghodss/yaml"
)

// ConvertToV1Config converts a stored pipelineModel.Yaml into a
// v1yaml.Config by round-tripping through JSON.
func ConvertToV1Config(src *pipelineModel.Yaml) (*v1yaml.Config, error) {
	if src == nil {
		return nil, fmt.Errorf("source yaml config is nil")
	}
	// Marshal pipelineModel.Yaml to JSON
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yaml config: %w", err)
	}
	// Unmarshal JSON to v1yaml.Config
	dst := &v1yaml.Config{}
	if err := json.Unmarshal(data, dst); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to v1yaml.Config: %w", err)
	}
	return dst, nil
}

// collectPlugins walks the pipeline config, resolves every referenced
// step plugin and returns a map of plugin name -> definition (YAML string).
//
// Resolution is best-effort: missing or broken plugins are skipped with a
// warning so the stage Detail request never fails; the runner will fail
// with a clear error when a referenced plugin definition is missing.
func (m *Manager) collectPlugins(ctx context.Context, data []byte) map[string]string {
	config, err := v1yaml.ParseBytes(data)
	if err != nil {
		logger.GetLogger(ctx).Warnw("collectPlugins: cannot parse pipeline config", "error", err)
		return nil
	}

	names := map[string]struct{}{}
	if err := walk.Walk(config, func(node interface{}) error {
		step, ok := node.(*v1yaml.Step)
		if !ok {
			return nil
		}
		p, ok := step.Spec.(*v1yaml.StepPlugin)
		if !ok {
			return nil
		}
		name := p.Name
		if p.Uses != "" {
			name = p.Uses
		}
		if name != "" {
			names[name] = struct{}{}
		}
		return nil
	}); err != nil {
		logger.GetLogger(ctx).Warnw("collectPlugins: walk failed", "error", err)
		return nil
	}

	plugins := map[string]string{}
	for name := range names {
		yamlCfg, err := m.PipelineSvc.Resolve(ctx, "plugin", "step", name)
		if err != nil {
			logger.GetLogger(ctx).Warnw("collectPlugins: cannot resolve plugin", "plugin", name, "error", err)
			continue
		}
		v1, err := ConvertToV1Config(yamlCfg)
		if err != nil {
			logger.GetLogger(ctx).Warnw("collectPlugins: cannot convert plugin", "plugin", name, "error", err)
			continue
		}
		b, err := yaml.Marshal(v1)
		if err != nil {
			logger.GetLogger(ctx).Warnw("collectPlugins: cannot marshal plugin", "plugin", name, "error", err)
			continue
		}
		plugins[name] = string(b)
	}
	return plugins
}
