package manager

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/logger"
	v1yaml "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/pipeline/spec/parse"
	"github.com/easysoft/gitfox-community/pkg/pipeline/spec/parse/walk"

	"github.com/ghodss/yaml"
)

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
		v1, err := parse.ToConfig(yamlCfg)
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
