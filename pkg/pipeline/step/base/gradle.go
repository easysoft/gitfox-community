// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"strings"

	harness "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

const (
	envGradlePluginMirror = "GRADLE_PLUGIN_MIRROR_URL"
)

type buildGradleStep struct {
}

func (c *buildGradleStep) Convert(step *harness.Step) (*harness.StepBuild, error) {
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

func (c *buildGradleStep) setupCache(r *StepRequest) {
	if r.OutSpec.Caching.KeyFile == "" {
		r.OutSpec.Caching.KeyFile = "build.gradle"
	}
	if r.OutSpec.Caching.Paths == nil || len(r.OutSpec.Caching.Paths) == 0 {
		r.OutSpec.Caching.Paths = []string{".cache/gradle/caches", ".cache/gradle/native"}
	}

	r.OutSpec.Envs["GRADLE_USER_HOME"] = "/gitfox/.cache/gradle"
}

func (c *buildGradleStep) setupMirror(r *StepRequest) {
	if _, exist := r.OutSpec.Envs[envMavenMirror]; !exist {
		r.OutSpec.Envs[envMavenMirror] = defaultMavenMirror
		if len(r.InSpec.Mirror) > 0 {
			mirror := findMirrorByKey("", r.InSpec.Mirror)
			if mirror != "" {
				r.OutSpec.Envs[envMavenMirror] = mirror
			}
		}
	}

	if _, exist := r.OutSpec.Envs[envGradlePluginMirror]; !exist {
		r.OutSpec.Envs[envGradlePluginMirror] = "https://maven.aliyun.com/repository/gradle-plugin"
		if len(r.InSpec.Mirror) > 0 {
			mirror := findMirrorByKey("plugin", r.InSpec.Mirror)
			if mirror != "" {
				r.OutSpec.Envs[envGradlePluginMirror] = mirror
			}
		}
	}

	r.OutSpec.PreActions = append(r.OutSpec.PreActions, "setup-gradle")
}

func (c *buildGradleStep) setupContainer(r *StepRequest) {
	r.OutSpec.Container.Image = "hub.zentao.net/ci/gradle:8.14.3-jdk11-alpine"
}

func (c *buildGradleStep) setProxy(r *StepRequest) {
}

func findMirrorByKey(key string, mirrorList harness.Stringorslice) string {
	if mirrorList == nil || len(mirrorList) == 0 {
		return ""
	}

	for _, mirror := range mirrorList {
		if key == "" && !strings.Contains(mirror, ":") {
			return mirror
		}

		if key != "" {
			if strings.Contains(mirror, ":") {
				prefix := strings.Split(mirror, ":")[0]
				if prefix == key {
					return mirror[strings.Index(mirror, ":")+1:]
				}
			}
		}
	}

	return ""
}
