package resources

import "embed"

var (
	//go:embed locales
	EmbeddedLocales embed.FS

	//go:embed license
	Licence embed.FS

	//go:embed gitignore
	Gitignore embed.FS
)
