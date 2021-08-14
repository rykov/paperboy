// +build withUI

package ui

import (
	"embed"
	"io/fs"
)

//go:embed dist/* dist/assets/*
var distFS embed.FS

// Above FS without prefix
var FS fs.FS

// Strip "dist/" prefix
func init() {
	FS, _ = fs.Sub(distFS, "dist")
}
