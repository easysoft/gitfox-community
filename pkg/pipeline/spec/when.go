package yaml

import (
	"encoding/json"
	"errors"
)

type (
	// When defines execution conditions.
	When struct {
		Eval    string
		Cond    []map[string]*Expr
		Changes []string
	}

	// Expr defines an execution expression.
	Expr struct {
		Eq  string   `json:"eq,omitempty"`
		In  []string `json:"in,omitempty"`
		Not *Expr    `json:"not,omitempty"`
	}
)

// UnmarshalJSON implements the unmarshal interface.
func (v *When) UnmarshalJSON(data []byte) error {
	// First try to parse as a complex object that might contain changes field
	type whenTemp struct {
		Eval    string   `json:"eval,omitempty"`
		Changes []string `json:"changes,omitempty"`
	}
	var tmp whenTemp
	if err := json.Unmarshal(data, &tmp); err == nil {
		v.Eval = tmp.Eval
		v.Changes = tmp.Changes
		// If only changes field was set, return early
		if len(v.Changes) > 0 || v.Eval != "" {
			return nil
		}
	}

	// parse the expression string.
	if v.Eval == "" { // Only try to parse as string if not already parsed above
		if err := json.Unmarshal(data, &v.Eval); err == nil {
			return nil
		}
	} else if v.Eval != "" {
		// 如果已经解析到了Eval字段值，直接返回
		return nil
	}

	// parse the declarative when clause array.
	if err := json.Unmarshal(data, &v.Cond); err == nil {
		return nil
	}

	// parse the declarative when clause.
	vv := map[string]*Expr{}
	if err := json.Unmarshal(data, &vv); err == nil {
		v.Cond = append(v.Cond, vv)
		return nil
	}

	return errors.New("failed to unmarshal when clause")
}

// MarshalJSON implements the marshal interface.
func (v *When) MarshalJSON() ([]byte, error) {
	if len(v.Changes) > 0 {
		type whenTemp struct {
			Eval    string             `json:"eval,omitempty"`
			Cond    []map[string]*Expr `json:"cond,omitempty"`
			Changes []string           `json:"changes,omitempty"`
		}
		return json.Marshal(whenTemp{
			Eval:    v.Eval,
			Cond:    v.Cond,
			Changes: v.Changes,
		})
	}

	if v.Eval != "" {
		return json.Marshal(v.Eval)
	}
	if len(v.Cond) > 0 {
		return json.Marshal(v.Cond)
	}
	return []byte("null"), nil
}
