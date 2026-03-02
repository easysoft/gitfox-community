// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	harness "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
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
