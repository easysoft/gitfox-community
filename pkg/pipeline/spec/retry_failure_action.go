// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"fmt"
)

// Failure action on failure for all retries
type RetryFailureAction struct {
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *RetryFailureAction) UnmarshalJSON(data []byte) error {
	type S RetryFailureAction
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "success":
		v.Spec = new(EmptySpec)
	case "fail":
		v.Spec = new(EmptySpec)
	case "stage-rollback":
		v.Spec = new(EmptySpec)
	case "pipeline-rollback":
		v.Spec = new(EmptySpec)
	case "abort":
		v.Spec = new(Abort)
	case "ignore":
		v.Spec = new(Ignore)
	case "manual-intervention":
		v.Spec = new(ManualIntervention)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
