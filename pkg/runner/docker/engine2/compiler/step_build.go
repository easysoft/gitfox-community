// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"strings"

	"github.com/easysoft/gitfox-community/app/service/pipeline"
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/runner/docker/engine2/engine"
	"github.com/easysoft/gitfox-community/pkg/utils/common"

	"github.com/drone/runner-go/environ"
)

func createBuildSteps(src *harness.Step, spec *harness.StepBuild, svc *pipeline.Service) ([]*engine.Step, error) {
	var convertedSteps []*engine.Step

	buildSpec := &harness.StepRun{
		Shell:     spec.Shell,
		Script:    spec.Script,
		Container: spec.Container,
		Envs:      spec.Envs,
		Reports:   spec.Reports,
		Outputs:   spec.Outputs,
		Mount:     spec.Mount,
	}

	// ctx := context.Background()
	if spec.BuildEnv != "" {
		// env, err := svc.GetBuildEnv(ctx, spec.BuildEnv)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to get build env '%s': %w", spec.BuildEnv, err)
		// }
		// if env.Category != constant.BuildEnvCategoryInternal {
		// 	return nil, fmt.Errorf("only internal build env should be use here")
		// }

		// buildSpec.Container.Image = env.Image
		// buildSpec.Container.Pull = env.Policy.String()
	}

	buildStep := createRunStep(src, buildSpec)
	setupScripts(buildStep, buildSpec.Script, "linux")

	if preBuildStep := createPreBuildStep(src, spec); preBuildStep != nil {
		convertedSteps = append(convertedSteps, preBuildStep)
	}

	convertedSteps = append(convertedSteps, buildStep)

	if postBuildStep := createPostBuildStep(src, spec); postBuildStep != nil {
		convertedSteps = append(convertedSteps, postBuildStep)
	}

	for _, dst := range convertedSteps {
		dst.WorkingDir = "/gitfox"
	}

	return convertedSteps, nil
}

func createPreBuildStep(src *harness.Step, spec *harness.StepBuild) *engine.Step {

	if *spec.Caching.Enabled || len(spec.PreActions) > 0 {
		preBuildSrc := copyStep(src)
		preBuildSrc.Id = "pre-build-" + src.Id
		preBuildSrc.Name = src.Name + " - Pre Actions"
		preBuildSpec := &harness.StepRun{
			Shell:  spec.Shell,
			Script: make(harness.Stringorslice, 0),
			Container: &harness.Container{
				Image: common.HelperImage(),
				Pull:  "always",
			},
			Mount: spec.Mount,
		}

		if *spec.Caching.Enabled {
			preBuildSpec.Script = append(preBuildSpec.Script, "pre-cache")
			cachingEnvs := map[string]string{
				"BUILD_CACHE_KEY_FILE": spec.Caching.KeyFile,
				"BUILD_CACHE_DIRS":     strings.Join(spec.Caching.Paths, ","),
			}
			preBuildSpec.Envs = environ.Combine(spec.Envs, cachingEnvs)
		}
		if len(spec.PreActions) > 0 {
			preBuildSpec.Script = append(preBuildSpec.Script, spec.PreActions...)
		}

		preCacheStep := createRunStep(preBuildSrc, preBuildSpec)
		setupScripts(preCacheStep, preBuildSpec.Script, "linux")
		return preCacheStep
	}
	return nil
}

func createPostBuildStep(src *harness.Step, spec *harness.StepBuild) *engine.Step {
	if *spec.Caching.Enabled || len(spec.PostActions) > 0 {
		postBuildSrc := copyStep(src)
		postBuildSrc.Id = "post-build-" + src.Id
		postBuildSrc.Name = src.Name + " - Post Actions"
		postBuildSpec := &harness.StepRun{
			Shell:  spec.Shell,
			Script: make(harness.Stringorslice, 0),
			Container: &harness.Container{
				Image: common.HelperImage(),
				Pull:  "always",
			},
			Mount: spec.Mount,
		}
		if *spec.Caching.Enabled {
			postBuildSpec.Script = append(postBuildSpec.Script, "post-cache")
			cachingEnvs := map[string]string{
				"BUILD_CACHE_KEY_FILE": spec.Caching.KeyFile,
				"BUILD_CACHE_DIRS":     strings.Join(spec.Caching.Paths, ","),
			}

			postBuildSpec.Envs = environ.Combine(spec.Envs, cachingEnvs)
		}
		if len(spec.PostActions) > 0 {
			postBuildSpec.Script = append(postBuildSpec.Script, spec.PostActions...)
		}

		postCacheStep := createRunStep(postBuildSrc, postBuildSpec)
		setupScripts(postCacheStep, postBuildSpec.Script, "linux")
		return postCacheStep
	}
	return nil
}

func copyStep(in *harness.Step) *harness.Step {
	var out harness.Step
	out = *in
	return &out
}
