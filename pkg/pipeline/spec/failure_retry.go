package yaml

type Retry struct {
	Attempts int64           `json:"attempts,omitempty"`
	Interval Durationorslice `json:"interval,omitempty"`
	Failure  *RetryFailure   `json:"failure,omitempty"`
}
