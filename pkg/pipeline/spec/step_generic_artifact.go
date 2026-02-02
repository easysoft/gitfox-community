package yaml

type StepGenericArtifact struct {
	SourcePath Stringorslice     `json:"sourcePath,omitempty"`
	SpaceRef   string            `json:"spaceRef,omitempty"`
	GroupId    int64             `json:"groupId,omitempty"`
	Envs       map[string]string `json:"envs,omitempty"`
	Container  *Container        `json:"container,omitempty"`
}
