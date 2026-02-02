package yaml

// Defines a matrix execution strategy.
type Matrix struct {
	Axis        map[string][]string `json:"axis,omitempty"`
	Concurrency int64               `json:"concurrency,omitempty"`
	Include     []map[string]string `json:"include,omitempty"`
	Exclude     []map[string]string `json:"exclude,omitempty"`
}
