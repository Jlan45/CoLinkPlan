package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// GetUIFS returns a filesystem localized to the dist folder
func GetUIFS() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
