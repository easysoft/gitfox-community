package yaml

type StepQltyScan struct {
	ScanPlanID int64             `json:"scanPlanID,omitempty"`
	Container  *Container        `json:"container,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
	Reports    []*Report         `json:"reports,omitempty"`
	Outputs    []string          `json:"outputs,omitempty"`
	Mount      []*Mount          `json:"mount,omitempty"`
}
