// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"fmt"

	harness "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/runner/docker/engine2/engine"
)

type state struct {
	options   *harness.Default
	platform  *harness.Platform
	resources *harness.Resources
	labels    map[string]string
	envs      map[string]string
}

func (c *CompilerImpl) convertStep(stage *harness.Stage, step *harness.Step, args Args) ([]*engine.Step, error) {
	switch v := step.Spec.(type) {
	case *harness.StepScript:
		dst := createStep(step, v)
		dst.WorkingDir = "/gitfox"
		setupScript(dst, v.Run, "linux")
		return []*engine.Step{dst}, nil
	case *harness.StepRun:
		dst := createRunStep(step, v)
		dst.WorkingDir = "/gitfox"
		setupScripts(dst, v.Script, "linux")
		return []*engine.Step{dst}, nil
	case *harness.StepPlugin:
		dst := createStepPlugin(step, v)
		return []*engine.Step{dst}, nil
	case *harness.StepBackground:
		dst := createStepBackground(step, v)
		setupScript(dst, v.Run, "linux")
		return []*engine.Step{dst}, nil
	}
	return nil, fmt.Errorf("unknown step type: %s", step.Type)
}
