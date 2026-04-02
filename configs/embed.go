package configs

import (
	"embed"
)

// FS contains embedded configuration files.
//
//go:embed prod.env
var FS embed.FS
