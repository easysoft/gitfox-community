// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package compiler

import (
	"fmt"
	"strings"

	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/runner/docker/engine2/engine"
	"github.com/easysoft/gitfox-community/pkg/runner/docker/internal/docker/image"
	"github.com/easysoft/gitfox-community/pkg/utils/common"

	"github.com/drone/runner-go/environ"
)

func createRawArtifactStep(src *harness.Step, spec *harness.StepRawArtifact) *engine.Step {
	dst := &engine.Step{
		ID:         random(),
		Name:       src.Id,
		Image:      common.HelperImage(),
		Command:    nil,
		Entrypoint: nil,
		// TODO re-enable
		//DependsOn:    spec.DependsOn,
		// DNS:          spec.DNS,
		// TODO re-enable
		// DNSSearch:    spec.DNSSearch,
		Envs: spec.Envs,
		// TODO re-enable
		// ExtraHosts:   spec.ExtraHosts,
		IgnoreStderr: false,
		IgnoreStdout: false,

		Networks: nil, // set in compiler.go
		Volumes:  nil, // set below
		Devices:  nil, // see below

		Display: src.Name,
	}

	if container := spec.Container; container != nil {
		dst.Image = common.GetImageOrDefault(image.Expand(container.Image), dst.Image)
		dst.Privileged = container.Privileged
		dst.Pull = convertPullPolicy(container.Pull)
	}

	if dst.Envs == nil {
		dst.Envs = map[string]string{}
	}

	// append all matrix parameters as environment
	// variables into the step
	if src.Strategy != nil && src.Strategy.Spec != nil {
		v, ok := src.Strategy.Spec.(*harness.Matrix)
		if ok {
			for _, axis := range v.Include {
				dst.Envs = environ.Combine(dst.Envs, axis)
			}
		}
	}

	return dst
}

func genRawArtifactScript(spec *harness.StepRawArtifact) string {
	cmd := []string{"uploader", "--format", "raw", "--space", spec.SpaceRef, "--name", spec.Name}

	if spec.Group != "" {
		cmd = append(cmd, "--group", spec.Group)
	}

	cmd = append(cmd, "--version", spec.Version)
	for _, pattern := range spec.SourcePath {
		cmd = append(cmd, "--file", fmt.Sprintf("'%s'", pattern))
	}

	return strings.Join(cmd, " ")
}
