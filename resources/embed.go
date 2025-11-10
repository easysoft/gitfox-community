package resources

import "embed"

var (
	//go:embed locales
	EmbeddedLocales embed.FS

	//go:embed license
	licence embed.FS

	//go:embed gitignore
	gitignore embed.FS
)
