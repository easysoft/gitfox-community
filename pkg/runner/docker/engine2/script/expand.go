// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package script

import (
	"regexp"
	"strings"

	schema "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
)

var pattern = regexp.MustCompile(`\${{(.*?)}}`)

// Expand expands the script inside the text snippet.
func Expand(code string, inputs map[string]interface{}) string {
	if !strings.Contains(code, "${{") {
		return code
	}
	return pattern.ReplaceAllStringFunc(code, func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "${{")
		s = strings.TrimSuffix(s, "}}")
		out, _ := EvalStr(s, inputs)
		return out
	})
}

// ExpandConfig expands scripts in the stage.
func ExpandConfig(in *schema.Config, inputs map[string]interface{}) {
	switch v := in.Spec.(type) {
	case *schema.Pipeline:
		for _, vv := range v.Stages {
			ExpandStage(vv, inputs)
		}
	case *schema.PluginStep:
	case *schema.PluginStage:
	case *schema.TemplateStage:
	case *schema.TemplateStep:
	}
}

// ExpandStage expands scripts in the stage.
func ExpandStage(stage *schema.Stage, inputs map[string]interface{}) {
	stage.Id = Expand(stage.Id, inputs)
	stage.Name = Expand(stage.Name, inputs)
	for i, s := range stage.Delegate {
		stage.Delegate[i] = Expand(s, inputs)
	}

	if stage.Strategy != nil && stage.Strategy.Spec != nil {
		if matrix, ok := stage.Strategy.Spec.(*schema.Matrix); ok {
			for _, axis := range matrix.Include {
				// FIXME this prevents matrix inside a matrix
				// FIXME ideally we aren't mutating the inputs here and make a copy instead
				inputs["matrix"] = axis
			}
		}
	}

	// FIXME ideally we aren't mutating the inputs and are making a copy instead
	for k, v := range stage.Inputs {
		m, ok := inputs["inputs"]
		if ok {
			mm, ok := m.(map[string]interface{})
			if ok {
				mm[k] = v
			}
		}
	}

	switch spec := stage.Spec.(type) {
	case *schema.StageCI:

		if vv := spec.Runtime; vv != nil {
			// switch vv.Spec.(type) {
			// case *schema.RuntimeCloud: // TODO
			// case *schema.RuntimeMachine: // TODO
			// case *schema.RuntimeVM: // TODO
			// case *schema.RuntimeKube: // TODO
			// }
		}

		// TODO
		if vv := spec.Platform; vv != nil {
			// vv.Os
			// vv.Arch
		}

		// TODO
		if vv := spec.Clone; vv != nil {
			// TODO
		}

		if vv := spec.Cache; vv != nil {
			vv.Key = Expand(vv.Key, inputs)
			vv.Policy = Expand(vv.Key, inputs)
			for i, s := range vv.Paths {
				vv.Paths[i] = Expand(s, inputs)
			}
			vv.Key = Expand(vv.Key, inputs)
		}

		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	}
}

// EexpandStep expands scripts in the step.
func ExpandStep(step *schema.Step, inputs map[string]interface{}) {
	step.Id = Expand(step.Id, inputs)
	step.Name = Expand(step.Name, inputs)

	if step.Strategy != nil && step.Strategy.Spec != nil {
		if matrix, ok := step.Strategy.Spec.(*schema.Matrix); ok {
			for _, axis := range matrix.Include {
				// FIXME this prevents matrix inside a matrix
				// FIXME ideally we aren't mutating the inputs here and make a copy instead
				inputs["matrix"] = axis
			}
		}
	}

	// FIXME ideally we aren't mutating the inputs and are making a copy instead
	for k, v := range step.Inputs {
		m, ok := inputs["inputs"]
		if ok {
			mm, ok := m.(map[string]interface{})
			if ok {
				if s, isstr := v.(string); isstr {
					// if the pipeline defines a template or plugin,
					// and uses an expression as the input, expand now
					mm[k] = Expand(s, inputs)
				} else {
					mm[k] = v
				}
			}
		}
	}

	switch spec := step.Spec.(type) {
	case *schema.StepBackground:
		spec.Run = Expand(spec.Run, inputs)
		spec.Image = Expand(spec.Image, inputs)
		spec.Entrypoint = Expand(spec.Entrypoint, inputs)
		for i, s := range spec.Args {
			spec.Args[i] = Expand(s, inputs)
		}
	case *schema.StepScript:
		spec.Run = Expand(spec.Run, inputs)
		spec.Image = Expand(spec.Image, inputs)
		spec.Connector = Expand(spec.Connector, inputs)
		spec.Entrypoint = Expand(spec.Entrypoint, inputs)
		for i, s := range spec.Args {
			spec.Args[i] = Expand(s, inputs)
		}
		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		if spec.Reports != nil {
			for _, report := range spec.Reports {
				for i, s := range report.Path {
					report.Path[i] = Expand(s, inputs)
				}
			}
		}

	case *schema.StepGroup:
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	case *schema.StepParallel:
		for _, vv := range spec.Steps {
			ExpandStep(vv, inputs)
		}
	case *schema.StepRun:
		for i, s := range spec.Script {
			spec.Script[i] = Expand(s, inputs)
		}
		for k, v := range spec.Envs {
			spec.Envs[k] = Expand(v, inputs)
		}
		if spec.Reports != nil {
			for _, report := range spec.Reports {
				for i, s := range report.Path {
					report.Path[i] = Expand(s, inputs)
				}
			}
		}
		if container := spec.Container; container != nil {
			container.Image = Expand(container.Image, inputs)
			container.Connector = Expand(container.Connector, inputs)
			container.Entrypoint = Expand(container.Entrypoint, inputs)
			for i, s := range container.Args {
				container.Args[i] = Expand(s, inputs)
			}
		}
	case *schema.StepPlugin:
		// NOTE
		// this case should never happen. we find and replace
		// all plugins with the plugin steps.
	case *schema.StepTemplate:
		// NOTE
		// this case should never happen. we find and replace
		// all templates with the template steps.
	}
}
