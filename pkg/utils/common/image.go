package common

func GitImage() string {
	return GetEnv("PIPELINE_GIT_IMAGE", DefaultGitImage)
}

func PlaceholderImage() string {
	return GetEnv("PIPELINE_PLACEHOLDER_IMAGE", DefaultPlaceholderImage)
}

func HelperImage() string {
	return GetEnv("PIPELINE_HELPER_IMAGE", DefaultPipelineHelperImage)
}

func GetImageOrDefault(in, defaults string) string {
	if in == "" {
		return defaults
	}
	return in
}
