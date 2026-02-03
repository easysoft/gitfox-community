package yaml

type StepRawArtifact struct {
	SourcePath Stringorslice     `json:"sourcePath,omitempty"`
	Name       string            `json:"name,omitempty"`
	Group      string            `json:"group,omitempty"`
	Version    string            `json:"version,omitempty"`
	SpaceRef   string            `json:"spaceRef,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
	Container  *Container        `json:"container,omitempty"`
}
