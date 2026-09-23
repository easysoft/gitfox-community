// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"errors"
)

type FailureList struct {
	Items []*Failure `json:"failure,omitempty"`
}

// UnmarshalJSON implements the unmarshal interface.
func (v *FailureList) UnmarshalJSON(data []byte) error {
	// parse the failure clause array.
	if err := json.Unmarshal(data, &v.Items); err == nil {
		return nil
	}

	// parse the simple failure clause.
	vv := new(Failure)
	if err := json.Unmarshal(data, &vv); err == nil {
		v.Items = append(v.Items, vv)
		return nil
	}

	return errors.New("failed to unmarshal failure clause")
}

// MarshalJSON implements the marshal interface.
func (v *FailureList) MarshalJSON() ([]byte, error) {
	if len(v.Items) == 1 {
		return json.Marshal(v.Items[0])
	}
	if len(v.Items) > 1 {
		return json.Marshal(v.Items)
	}
	return []byte("null"), nil
}
