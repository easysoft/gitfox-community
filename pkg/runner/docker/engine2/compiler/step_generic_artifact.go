// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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

func createGenericArtifactStep(src *harness.Step, spec *harness.StepGenericArtifact) *engine.Step {
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

func genGenericArtifactScript(spec *harness.StepGenericArtifact) string {
	cmd := []string{"uploader", "--format", "generic", "--space", spec.SpaceRef, "--group", fmt.Sprintf("%d", spec.GroupId)}

	for _, pattern := range spec.SourcePath {
		cmd = append(cmd, "--file", fmt.Sprintf("'%s'", pattern))
	}

	return strings.Join(cmd, " ")
}
