// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package manager

import (
	"context"
	"errors"
	"strings"
	"testing"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/app/service/pipeline"
)

// fakePipelineSvc implements pipeline.Interface via embedding; only
// Resolve is overridden for the tests below.
type fakePipelineSvc struct {
	pipeline.Interface
	resolve func(ctx context.Context, kind, typ, name string) (*pipelineModel.Yaml, error)
}

func (f *fakePipelineSvc) Resolve(ctx context.Context, kind, typ, name string) (*pipelineModel.Yaml, error) {
	return f.resolve(ctx, kind, typ, name)
}

func pluginYaml(name string) *pipelineModel.Yaml {
	return &pipelineModel.Yaml{
		Kind: "plugin",
		Type: pipelineModel.PluginTypeStep,
		Name: name,
		Spec: map[string]interface{}{
			"name": name,
			"step": map[string]interface{}{
				"type": "script",
				"spec": map[string]interface{}{
					"entry": "gffox clone --depth 1",
					"image": "hub.zentao.net/ci/gitfox-plugin-git:gitfox-community",
					"envs": map[string]interface{}{
						"PLUGIN_CLONE_REPOSITORY": "${{ inputs.repository}}",
					},
				},
			},
			"inputs": map[string]interface{}{
				"repository": map[string]interface{}{
					"type":        "string",
					"description": "remote url",
					"default":     "https://example.com/repo.git",
				},
			},
		},
	}
}

func TestCollectPlugins(t *testing.T) {
	m := &Manager{PipelineSvc: &fakePipelineSvc{
		resolve: func(ctx context.Context, kind, typ, name string) (*pipelineModel.Yaml, error) {
			if name == "git_clone" {
				return pluginYaml("git_clone"), nil
			}
			return nil, errors.New("plugin not found")
		},
	}}

	data := []byte(`---
version: 1
kind: pipeline
spec:
  stages:
    - name: build
      type: ci
      spec:
        steps:
          - type: plugin
            spec:
              name: git_clone
`)

	plugins := m.collectPlugins(context.Background(), data)
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d: %v", len(plugins), plugins)
	}
	def, ok := plugins["git_clone"]
	if !ok {
		t.Fatalf("missing plugin git_clone: %v", plugins)
	}
	if !strings.Contains(def, "gffox clone --depth 1") {
		t.Errorf("plugin definition missing entry: %s", def)
	}
}

func TestCollectPluginsUses(t *testing.T) {
	m := &Manager{PipelineSvc: &fakePipelineSvc{
		resolve: func(ctx context.Context, kind, typ, name string) (*pipelineModel.Yaml, error) {
			return pluginYaml(name), nil
		},
	}}

	data := []byte(`---
version: 1
kind: pipeline
spec:
  stages:
    - name: build
      type: ci
      spec:
        steps:
          - type: plugin
            spec:
              uses: git_clone
`)

	plugins := m.collectPlugins(context.Background(), data)
	if _, ok := plugins["git_clone"]; !ok {
		t.Fatalf("expected plugin git_clone via uses, got %v", plugins)
	}
}

func TestCollectPluginsSkipsUnresolvable(t *testing.T) {
	m := &Manager{PipelineSvc: &fakePipelineSvc{
		resolve: func(ctx context.Context, kind, typ, name string) (*pipelineModel.Yaml, error) {
			return nil, errors.New("plugin not found")
		},
	}}

	data := []byte(`---
version: 1
kind: pipeline
spec:
  stages:
    - name: build
      type: ci
      spec:
        steps:
          - type: plugin
            spec:
              name: missing_plugin
`)

	plugins := m.collectPlugins(context.Background(), data)
	if len(plugins) != 0 {
		t.Fatalf("expected empty plugins for unresolvable refs, got %v", plugins)
	}
}

func TestCollectPluginsNone(t *testing.T) {
	m := &Manager{PipelineSvc: &fakePipelineSvc{}}

	data := []byte(`---
version: 1
kind: pipeline
spec:
  stages:
    - name: build
      type: ci
      spec:
        steps:
          - type: exec
            spec:
              script:
                - echo hello
`)

	plugins := m.collectPlugins(context.Background(), data)
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins, got %v", plugins)
	}
}

func TestCollectPluginsInvalidConfig(t *testing.T) {
	m := &Manager{PipelineSvc: &fakePipelineSvc{}}
	plugins := m.collectPlugins(context.Background(), []byte("not: [valid: yaml"))
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins for invalid config, got %v", plugins)
	}
}
