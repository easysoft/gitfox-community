// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"fmt"
)

type Step struct {
	Id       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Desc     string                 `json:"desc,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Timeout  string                 `json:"timeout,omitempty"`
	Strategy *Strategy              `json:"strategy,omitempty"`
	When     *When                  `json:"when,omitempty"`
	Failure  *FailureList           `json:"failure,omitempty"`
	Inputs   map[string]interface{} `json:"inputs,omitempty"`
	Spec     interface{}            `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Step) UnmarshalJSON(data []byte) error {
	type S Step
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "build":
		v.Spec = new(StepBuild)
	case "build_go":
		v.Spec = new(StepBuildCommon)
	case "build_yarn":
		v.Spec = new(StepBuildCommon)
	case "build_npm":
		v.Spec = new(StepBuildCommon)
	case "build_maven":
		v.Spec = new(StepBuildCommon)
	case "build_gradle":
		v.Spec = new(StepBuildCommon)
	case "background":
		v.Spec = new(StepBackground)
	case "exec":
		v.Spec = new(StepExec)
	case "script":
		v.Spec = new(StepScript)
	case "run":
		v.Spec = new(StepRun)
	case "group":
		v.Spec = new(StepGroup)
	case "parallel":
		v.Spec = new(StepParallel)
	case "plugin":
		v.Spec = new(StepPlugin)
	case "template":
		v.Spec = new(StepTemplate)
	case "generic_artifact":
		v.Spec = new(StepGenericArtifact)
	case "raw_artifact":
		v.Spec = new(StepRawArtifact)
	case "win_qt_build":
		v.Spec = new(StepWinQtBuild)
	case "win_vs_build":
		v.Spec = new(StepWinVsBuild)
	case "win_package":
		v.Spec = new(StepWinPackage)
	case "asset_download":
		v.Spec = new(StepAssetDownload)
	case "docker_build":
		v.Spec = new(StepBuildAndPushDockerImage)
	case "qlty_scan":
		v.Spec = new(StepQltyScan)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
