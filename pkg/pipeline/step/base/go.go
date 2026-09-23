// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package base

import (
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

type buildGoStep struct {
}

func (c *buildGoStep) Convert(step *harness.Step) (*harness.StepBuild, error) {
	req, err := NewEmptyBuildStep(step)
	if err != nil {
		return nil, err
	}

	c.setupMirror(req)
	if getOrDefaultTrue(req.OutSpec.Caching.Enabled) {
		c.setupCache(req)
	}
	c.setupContainer(req)
	return req.OutSpec, nil
}

func (c *buildGoStep) setupCache(r *StepRequest) {
	if r.OutSpec.Caching.KeyFile == "" {
		r.OutSpec.Caching.KeyFile = "go.mod"
	}
	if r.OutSpec.Caching.Paths == nil || len(r.OutSpec.Caching.Paths) == 0 {
		r.OutSpec.Caching.Paths = []string{".cache/go"}
	}

	r.OutSpec.Envs["GOMODCACHE"] = "/gitfox/.cache/go/mod"
}

func (c *buildGoStep) setupMirror(r *StepRequest) {
	if _, exist := r.OutSpec.Envs["GOPROXY"]; !exist {
		r.OutSpec.Envs["GOPROXY"] = "https://goproxy.cn,direct"
		if len(r.InSpec.Mirror) > 0 {
			r.OutSpec.Envs["GOPROXY"] = r.InSpec.Mirror[0]
		}
	}
}

func (c *buildGoStep) setupContainer(r *StepRequest) {
	r.OutSpec.Container.Image = "hub.zentao.net/ci/golang:1.24.5-alpine3.22"
}
