package yaml

import "time"

type ManualIntervention struct {
	TimeoutAction *TimeoutAction `json:"timeout_action,omitempty"`
	Timeout       time.Duration  `json:"timeout,omitempty"`
}
