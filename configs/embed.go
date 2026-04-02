package configs

import (
	"embed"
)

// FS contains embedded configuration files.
//
//go:embed prod.env prod.yaml
var FS embed.FS
