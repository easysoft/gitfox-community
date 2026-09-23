// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package base

import (
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

type buildNpmStep struct {
}

func (c *buildNpmStep) Convert(step *harness.Step) (*harness.StepBuild, error) {
	req, err := NewEmptyBuildStep(step)
	if err != nil {
		return nil, err
	}

	c.setupMirror(req)
	if getOrDefaultTrue(req.OutSpec.Caching.Enabled) {
		c.setupCache(req)
	}
	c.setProxy(req)
	c.setupContainer(req)
	return req.OutSpec, nil
}

func (c *buildNpmStep) setupCache(r *StepRequest) {
	if r.OutSpec.Caching.KeyFile == "" {
		r.OutSpec.Caching.KeyFile = "package-lock.json"
	}
	if r.OutSpec.Caching.Paths == nil || len(r.OutSpec.Caching.Paths) == 0 {
		r.OutSpec.Caching.Paths = []string{".cache/npm"}
	}

	r.OutSpec.Envs["NPM_CONFIG_CACHE"] = "/gitfox/.cache/npm"
}

func (c *buildNpmStep) setupMirror(r *StepRequest) {
	mirror := "https://registry.npmmirror.com"
	if len(r.InSpec.Mirror) > 0 {
		mirror = r.InSpec.Mirror[0]
	}
	r.OutSpec.Envs["NPM_CONFIG_REGISTRY"] = mirror
}

func (c *buildNpmStep) setupContainer(r *StepRequest) {
	r.OutSpec.Container.Image = "hub.zentao.net/ci/node:20.11.1-bookworm-slim"
}

func (c *buildNpmStep) setProxy(r *StepRequest) {
}
