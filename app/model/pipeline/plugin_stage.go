package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TemplateStage defines a stage-level template
type TemplateStage struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Stage       *Stage            `json:"stage,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}

// PluginStage defines a stage-level plugin
type PluginStage struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Stage       *Stage            `json:"stage,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}

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

// Stringorslice represents
// Using engine-api Strslice and augment it with YAML marshalling stuff. a string or an array of strings.
type Stringorslice []string

// UnmarshalJSON implements the unmarshal interface.
func (s *Stringorslice) UnmarshalJSON(data []byte) error {
	var stringType string
	if err := json.Unmarshal(data, &stringType); err == nil {
		*s = []string{stringType}
		return nil
	}

	var sliceType []interface{}
	if err := json.Unmarshal(data, &sliceType); err == nil {
		parts, err := toStrings(sliceType)
		if err != nil {
			return err
		}
		*s = parts
		return nil
	}

	return errors.New("failed to unmarshal string or string array")
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("cannot unmarshal %v of type %T into a string value", v, v)
		}
	}
	return r, nil
}

type StatusStage struct {
	Disabled bool   `json:"disabled,omitempty"`
	Matrix   string `json:"matrix,omitempty"`
	Name     string `json:"name,omitempty"`
}
