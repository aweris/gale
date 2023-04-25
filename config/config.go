package config

import (
	"github.com/adrg/xdg"
	"path/filepath"
)

// DataHome returns the path to the data directory.
func DataHome() string {
	return filepath.Join(xdg.DataHome, "gale")
}

// SearchDataFile returns the path to the data file.
func SearchDataFile(relPath string) (string, error) {
	return xdg.SearchDataFile(filepath.Join("gale", relPath))
}
