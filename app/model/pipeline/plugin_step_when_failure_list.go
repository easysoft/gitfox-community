// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type FailureList struct {
	Items []*Failure `json:"failure,omitempty"`
}

type Failure struct {
	Errors []string       `json:"errors,omitempty"`
	Action *FailureAction `json:"action,omitempty"`
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

type EmptySpec struct {
}

type Abort struct {
}

type Ignore struct {
}

type Retry struct {
	Attempts int64           `json:"attempts,omitempty"`
	Interval Durationorslice `json:"interval,omitempty"`
	Failure  *RetryFailure   `json:"failure,omitempty"`
}

// Durationorslice represents a duration string or an array
// of duration strings.
type Durationorslice []time.Duration

// UnmarshalJSON implements the unmarshal interface.
func (s *Durationorslice) UnmarshalJSON(data []byte) error {
	var stringType time.Duration
	if err := json.Unmarshal(data, &stringType); err == nil {
		*s = []time.Duration{stringType}
		return nil
	}

	var sliceType []interface{}
	if err := json.Unmarshal(data, &sliceType); err == nil {
		parts, err := toDurations(sliceType)
		if err != nil {
			return err
		}
		*s = parts
		return nil
	}

	return errors.New("failed to unmarshal duration string or duration array")
}

func toDurations(s []interface{}) ([]time.Duration, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]time.Duration, len(s))
	for k, i := range s {
		switch v := i.(type) {
		case string:
			dur, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("cannot unmarshal %v into type duration", i)
			}
			r[k] = dur
		case int:
			r[k] = time.Duration(int64(v))
		case int64:
			r[k] = time.Duration(v)
		default:
			return nil, fmt.Errorf("cannot unmarshal %v into type duration", i)
		}
	}
	return r, nil
}

type RetryFailure struct {
	Action *RetryFailureAction `json:"action,omitempty"`
}

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

type ManualIntervention struct {
	TimeoutAction *TimeoutAction `json:"timeout_action,omitempty"`
	Timeout       time.Duration  `json:"timeout,omitempty"`
}

// Manual intervention for timeout failure action
type TimeoutAction struct {
	Type string      `json:"type,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *TimeoutAction) UnmarshalJSON(data []byte) error {
	type S TimeoutAction
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
	default:
		return fmt.Errorf("unknown type %s", v.Type)
	}

	return json.Unmarshal(obj.Spec, v.Spec)
}
