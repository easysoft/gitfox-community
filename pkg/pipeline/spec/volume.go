package yaml

import (
	"encoding/json"
	"fmt"
)

type Volume struct {
	Name string      `json:"name,omitempty"`
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Volume) UnmarshalJSON(data []byte) error {
	type S Volume
	type T struct {
		*S
		Spec json.RawMessage `json:"spec"`
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	switch v.Type {
	case "host":
		v.Spec = new(VolumeHost)
	case "claim":
		v.Spec = new(VolumeClaim)
	case "temp":
		v.Spec = new(VolumeTemp)
	case "config-map":
		v.Spec = new(VolumeConfigMap)
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
