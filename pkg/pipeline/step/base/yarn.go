// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"fmt"
	"path/filepath"

	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

type buildYarnStep struct {
}

func (c *buildYarnStep) Convert(step *harness.Step) (*harness.StepBuild, error) {
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

func (c *buildYarnStep) setupCache(r *StepRequest) {
	if r.OutSpec.Caching.KeyFile == "" {
		r.OutSpec.Caching.KeyFile = "yarn.lock"
	}
	if r.OutSpec.Caching.Paths == nil || len(r.OutSpec.Caching.Paths) == 0 {
		r.OutSpec.Caching.Paths = []string{".cache/yarn"}
	}

	setCacheCmd := fmt.Sprintf("yarn config set cache-folder %s", filepath.Join("/gitfox", ".cache", "yarn"))
	r.OutSpec.Script = append(harness.Stringorslice{setCacheCmd}, r.OutSpec.Script...)
}

func (c *buildYarnStep) setupMirror(r *StepRequest) {
	mirror := "https://registry.npmmirror.com"
	if len(r.InSpec.Mirror) > 0 {
		mirror = r.InSpec.Mirror[0]
	}
	setMirrorCmd := fmt.Sprintf("yarn config set registry %s", mirror)
	r.OutSpec.Script = append(harness.Stringorslice{setMirrorCmd}, r.OutSpec.Script...)
}

func (c *buildYarnStep) setupContainer(r *StepRequest) {
	r.OutSpec.Container.Image = "hub.zentao.net/ci/node:20.11.1-bookworm-slim"
}

func (c *buildYarnStep) setProxy(r *StepRequest) {
}
