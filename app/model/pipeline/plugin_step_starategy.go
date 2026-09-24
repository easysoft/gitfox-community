// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"encoding/json"
	"fmt"
)

// Strategy defines an optional execution strategy.
type Strategy struct {
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Strategy) UnmarshalJSON(data []byte) error {
	type S Strategy
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "for":
		v.Spec = new(For)
	case "matrix":
		v.Spec = new(Matrix)
	case "while":
		v.Spec = new(While)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}

// Defines a for loop execution strategy.
type For struct {
	Concurrency int64 `json:"concurrency,omitempty"`
	Iterations  int64 `json:"iterations,omitempty"`
}

// Defines a matrix execution strategy.
type Matrix struct {
	Axis        map[string][]string `json:"axis,omitempty"`
	Concurrency int64               `json:"concurrency,omitempty"`
	Include     []map[string]string `json:"include,omitempty"`
	Exclude     []map[string]string `json:"exclude,omitempty"`
}

// Defines a while loop execution strategy.
type While struct {
	Condition  string `json:"condition,omitempty"`
	Iterations int64  `json:"iterations,omitempty"`
}
