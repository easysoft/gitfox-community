package yaml

import (
	"encoding/json"
	"fmt"
)

// Failure defines a failure strategy.
type FailureAction struct {
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *FailureAction) UnmarshalJSON(data []byte) error {
	type S FailureAction
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
	case "retry-step-group":
		v.Spec = new(EmptySpec)
	case "stage-rollback":
		v.Spec = new(EmptySpec)
	case "pipeline-rollback":
		v.Spec = new(EmptySpec)
	case "abort":
		v.Spec = new(Abort)
	case "ignore":
		v.Spec = new(Ignore)
	case "retry":
		v.Spec = new(Retry)
	case "manual-intervention":
		v.Spec = new(ManualIntervention)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
