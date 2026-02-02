package yaml

type Failure struct {
	Errors []string       `json:"errors,omitempty"`
	Action *FailureAction `json:"action,omitempty"`
}
