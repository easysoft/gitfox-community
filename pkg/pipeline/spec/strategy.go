package yaml

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
