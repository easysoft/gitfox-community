// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package base

import (
	"errors"

	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"

	"github.com/gotidy/ptr"
)

type StepRequest struct {
	InSpec  *harness.StepBuildCommon
	OutSpec *harness.StepBuild
}

func NewEmptyBuildStep(step *harness.Step) (*StepRequest, error) {
	in, ok := step.Spec.(*harness.StepBuildCommon)
	if !ok {
		return nil, errors.New("invalid step spec")
	}

	out := &harness.StepBuild{
		Script: in.Script,
		Mount:  in.Mount,
		Caching: &harness.Caching{
			Enabled: ptr.Bool(true),
		},
		BuildEnv: in.BuildEnv,
	}

	if out.Envs == nil {
		out.Envs = make(map[string]string)
	} else {
		out.Envs = in.Envs
	}

	if in.Container != nil {
		out.Container = in.Container
	} else {
		out.Container = new(harness.Container)
	}

	if in.Caching != nil {
		out.Caching.Enabled = ptr.Bool(getOrDefaultTrue(in.Caching.Enabled))
		if in.Caching.KeyFile != "" {
			out.Caching.KeyFile = in.Caching.KeyFile
		}
		if in.Caching.Paths != nil {
			out.Caching.Paths = in.Caching.Paths
		}
	}
	return &StepRequest{
		InSpec:  in,
		OutSpec: out,
	}, nil
}
