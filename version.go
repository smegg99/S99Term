// version.go

// Package s99term is the terminal render layer for smegg99 command-line
// programs.
package s99term

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

// VERSION is embedded so the version is written in one place.
//
//go:embed VERSION
var versionFile string

// Declared returns the version written in VERSION.
func Declared() string { return strings.TrimSpace(versionFile) }

// Version is the declared version, plus the commit it was built from when that
// is knowable.
func Version() string {
	rev, dirty, ok := revision()
	if !ok {
		return Declared()
	}
	return format(Declared(), rev, dirty)
}

// format is separate from Version because a test binary carries no VCS stamps,
// so the only way to check the shapes is to hand them in.
func format(declared, rev string, dirty bool) string {
	if rev == "" {
		return declared
	}
	if dirty {
		return declared + " (" + rev + ", dirty)"
	}
	return declared + " (" + rev + ")"
}

func revision() (rev string, dirty, ok bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false, false
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if len(rev) < 7 {
		return "", false, false
	}
	return rev[:7], dirty, true
}
