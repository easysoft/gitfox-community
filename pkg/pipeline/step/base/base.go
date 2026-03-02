// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"errors"

	harness "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"

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
