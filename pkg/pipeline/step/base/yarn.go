// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
