package yaml

import (
	"encoding/json"
	"fmt"
)

// Defines the runtime execution engine.
type Runtime struct {
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Runtime) UnmarshalJSON(data []byte) error {
	type S Runtime
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "default":
		v.Spec = new(RuntimeDefault)
	case "agent":
		v.Spec = new(RuntimeAgent)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
