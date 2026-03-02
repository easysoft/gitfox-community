// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"fmt"

	harness "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"
)

const (
	envMavenMirror          = "MAVEN_MIRROR_URL"
	defaultMavenMirror      = "https://maven.aliyun.com/repository/public"
	envMavenLocalRepository = "MAVEN_LOCAL_REPOSITORY_PATH"
)

type buildMavenStep struct {
}

func (c *buildMavenStep) Convert(step *harness.Step) (*harness.StepBuild, error) {
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

func (c *buildMavenStep) setupCache(r *StepRequest) {
	if r.OutSpec.Caching.KeyFile == "" {
		r.OutSpec.Caching.KeyFile = "pom.xml"
	}
	if r.OutSpec.Caching.Paths == nil || len(r.OutSpec.Caching.Paths) == 0 {
		r.OutSpec.Caching.Paths = []string{".cache/maven"}
	}

	r.OutSpec.Envs[envMavenLocalRepository] = "/gitfox/.cache/maven"
}

func (c *buildMavenStep) setupMirror(r *StepRequest) {
	copySettingsCmd := fmt.Sprintf("mkdir ~/.m2 && cp %s ~/.m2/settings.xml", ".gen_settings.xml")
	r.OutSpec.Script = append(harness.Stringorslice{copySettingsCmd}, r.OutSpec.Script...)
	if _, exist := r.OutSpec.Envs[envMavenMirror]; !exist {
		r.OutSpec.Envs[envMavenMirror] = defaultMavenMirror
		if len(r.InSpec.Mirror) > 0 {
			r.OutSpec.Envs[envMavenMirror] = r.InSpec.Mirror[0]
		}
	}
	r.OutSpec.PreActions = append(r.OutSpec.PreActions, "setup-maven")
}

func (c *buildMavenStep) setupContainer(r *StepRequest) {
	r.OutSpec.Container.Image = "hub.zentao.net/ci/maven:3.9.4-eclipse-temurin-11-alpine"
}

func (c *buildMavenStep) setProxy(r *StepRequest) {
}
