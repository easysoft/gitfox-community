package yaml

// Defines a while loop execution strategy.
type While struct {
	Condition  string `json:"condition,omitempty"`
	Iterations int64  `json:"iterations,omitempty"`
}
