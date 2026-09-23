// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"fmt"
)

type Stage struct {
	Desc         string                 `json:"desc,omitempty"`
	Id           string                 `json:"id,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Strategy     *Strategy              `json:"strategy,omitempty"`
	Delegate     Stringorslice          `json:"delegate,omitempty"`
	Status       *StatusStage           `json:"status,omitempty"`
	Type         string                 `json:"type,omitempty"`
	When         *When                  `json:"when,omitempty"`
	ChangedFiles []string               `json:"changed_files,omitempty"`
	Failure      *FailureList           `json:"failure,omitempty"`
	Inputs       map[string]interface{} `json:"inputs,omitempty"`
	Spec         interface{}            `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Stage) UnmarshalJSON(data []byte) error {
	type S Stage
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "ci":
		v.Spec = new(StageCI)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
