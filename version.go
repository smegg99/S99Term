// version.go

// Package s99term is the terminal render layer for smegg99 command-line
// programs.
package s99term

import (
	_ "embed"
	"strings"
)

// VERSION is embedded so the version is written in one place.
//
//go:embed VERSION
var versionFile string

// Declared returns the version written in VERSION.
func Declared() string { return strings.TrimSpace(versionFile) }
