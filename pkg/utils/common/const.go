package common

const (
	// DefaultPipelineExecutionCreatedUnix is the default unix timestamp for pipeline execution creation time. 2024-03-01 00:00:00
	DefaultPipelineExecutionCreatedUnix int64 = 1709222400000

	// DefaultGitImage is the default git image
	DefaultGitImage = "hub.zentao.net/ci/gitfox-plugin-git:2.1"
	// DefaultPlaceholderImage is the default placeholder image
	DefaultPlaceholderImage = "hub.zentao.net/ci/gitfox-plugin-placeholder:1"

	DefaultGitfoxUploaderImage = "hub.zentao.net/ci/gitfox-uploader:latest"

	DefaultPipelineHelperImage = "hub.zentao.net/ci/pipeline-helper:latest"
)
