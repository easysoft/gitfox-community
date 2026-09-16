// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/app/service/pipeline"
	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/runner/docker/engine2/engine"
)

func createDockerBuildSteps(stage *harness.Stage, src *harness.Step,
	spec *harness.StepBuildAndPushDockerImage,
	args Args,
	svc *pipeline.Service,
) []*engine.Step {
	var convertedSteps []*engine.Step

	if spec.Context == "" {
		spec.Context = "."
	}

	if spec.Dockerfile == "" {
		spec.Dockerfile = "Dockerfile"
	}

	if len(spec.Platforms) == 0 {
		spec.Platforms = []string{"linux/amd64"}
	}

	convertedSteps = createDockerBuildDindStep(stage, src, spec, args, svc)

	return convertedSteps
}

func createDockerBuildDindStep(stage *harness.Stage, src *harness.Step, spec *harness.StepBuildAndPushDockerImage,
	args Args, svc *pipeline.Service) []*engine.Step {
	volPath := "/var/run"
	volName := "socketDir"

	stageSpec, ok := stage.Spec.(*harness.StageCI)
	if !ok {
		return nil
	}

	stageSpec.Volumes = append(stageSpec.Volumes, &harness.Volume{
		Name: volName,
		Type: "temp",
		Spec: &harness.VolumeTemp{},
	})

	bgSpec := &harness.StepBackground{
		Privileged: true,
		Image:      "hub.qucheng.com/ci/docker:24.0-dind",
		Mount: []*harness.Mount{
			{Name: volName, Path: volPath},
		},
	}
	bgSrc := copyStep(src)
	bgSrc.Id = "dind-" + src.Id
	bgSrc.Name = src.Name + " - daemon"
	bgStep := createStepBackground(bgSrc, bgSpec)

	runSpec := &harness.StepRun{
		Script: harness.Stringorslice{},
		Container: &harness.Container{
			Image: "hub.zentao.net/ci/docker-cli:27.3.1",
		},
		Mount: []*harness.Mount{
			{Name: volName, Path: volPath},
		},
	}

	runSpec.Script = append(runSpec.Script, "sleep 5")
	for _, loginCmd := range genDockerLoginCmds(spec, args, svc) {
		runSpec.Script = append(runSpec.Script, loginCmd)
	}

	runSpec.Script = append(runSpec.Script, "setup-buildx-container")

	runSpec.Script = append(runSpec.Script, genDockerBuildCmd(spec, args, svc))
	runStep := createRunStep(src, runSpec)
	setupScripts(runStep, runSpec.Script, "linux")
	runStep.WorkingDir = "/gitfox"

	return []*engine.Step{bgStep, runStep}
}

func genDockerBuildCmd(spec *harness.StepBuildAndPushDockerImage, args Args, svc *pipeline.Service) string {
	buildTags := genDockerTags(append(spec.ExtraRepositories, spec.Repository), args, svc)

	var buf bytes.Buffer
	buf.WriteString("docker buildx build --pull --push")
	if len(spec.Platforms) > 0 {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("--platform %s", strings.Join(spec.Platforms, ",")))
	}

	for _, tag := range buildTags {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("-t %s", tag))
	}

	for arg, val := range spec.BuildArgs {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("--build-arg %s=\"%s\"", arg, val))
	}

	for label, val := range spec.Labels {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("--label %s=\"%s\"", label, val))
	}

	buf.WriteString(" ")
	buf.WriteString(fmt.Sprintf("-f %s", spec.Dockerfile))

	buf.WriteString(" ")
	buf.WriteString(spec.Context)

	return buf.String()
}

func genDockerLoginCmds(spec *harness.StepBuildAndPushDockerImage, args Args, svc *pipeline.Service) []string {
	var cmds []string
	// for _, repo := range append(spec.ExtraRepositories, spec.Repository) {
	// 	if repo.ConnectorRef == "" {
	// 		cmds = append(cmds, "setup-login")
	// 		continue
	// 	}

	// 	r, err := svc.GetRegistry(context.Background(), args.Repo.Namespace, repo.ConnectorRef)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	cmd := fmt.Sprintf("setup-login -s %s -u %s -p ${{ secrets.get('%s')}}", r.Host, r.UserName, r.Password.Identifier)
	// 	cmds = append(cmds, cmd)
	// }
	return cmds
}

func genDockerTags(repositories []*harness.DockerRepository, args Args, svc *pipeline.Service) []string {
	var tags []string
	// for _, repo := range repositories {
	// 	var host string
	// 	if repo.ConnectorRef == "" {
	// 		host = "${GITFOX_SERVER}"
	// 	} else {
	// 		// todo: parse connect ref
	// 		r, err := svc.GetRegistry(context.Background(), args.Repo.Namespace, repo.ConnectorRef)
	// 		if err != nil {
	// 			continue
	// 		}
	// 		host = r.Host
	// 	}

	// 	for _, tag := range repo.Tags {
	// 		tags = append(tags, fmt.Sprintf("%s/%s:%s", host, repo.Name, tag))
	// 	}
	// }
	return tags
}
