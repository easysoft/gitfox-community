// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"fmt"
)

// Config defines the a resource configuration.
type Config struct {
	Version StringorInt `json:"version,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Type    string      `json:"type,omitempty"`
	Name    string      `json:"name,omitempty"`
	Spec    interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Config) UnmarshalJSON(data []byte) error {
	type S Config
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Kind {
	case "pipeline":
		v.Spec = new(Pipeline)
	case "template":
		switch v.Type {
		case "stage":
			v.Spec = new(TemplateStage)
		case "step":
			v.Spec = new(TemplateStep)
		default:
			return fmt.Errorf("unknown template type %s", v.Type)
		}
	case "plugin":
		switch v.Type {
		case "stage":
			v.Spec = new(PluginStage)
		case "step":
			v.Spec = new(PluginStep)
		default:
			return fmt.Errorf("unknown plugin type %s", v.Type)
		}
	default:
		return fmt.Errorf("unknown kind %s", v.Kind)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
