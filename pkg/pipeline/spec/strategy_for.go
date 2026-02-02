package yaml

// Defines a for loop execution strategy.
type For struct {
	Concurrency int64 `json:"concurrency,omitempty"`
	Iterations  int64 `json:"iterations,omitempty"`
}
