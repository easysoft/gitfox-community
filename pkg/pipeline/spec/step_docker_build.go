package yaml

type StepBuildAndPushDockerImage struct {
	Mode              string              `json:"mode,omitempty"`
	DriverOpt         string              `json:"driverOpt,omitempty"`
	Dockerfile        string              `json:"dockerfile,omitempty"`
	Context           string              `json:"context,omitempty"`
	Platforms         []string            `json:"platforms,omitempty"`
	BuildArgs         map[string]string   `json:"buildArgs,omitempty"`
	Labels            map[string]string   `json:"labels,omitempty"`
	Repository        *DockerRepository   `json:"repository,omitempty"`
	ExtraRepositories []*DockerRepository `json:"extraRepositories,omitempty"`
}
