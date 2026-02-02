package yaml

type StepWinQtBuild struct {
	Tool          string `json:"tool,omitempty"`
	Configuration string `json:"configuration,omitempty"`
	Arch          string `json:"arch,omitempty"`
	CustomFlags   string `json:"customFlags,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}
