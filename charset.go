// charset.go

package s99term

import "strings"

// utf8Vars is POSIX precedence: the first one set and non-empty decides.
var utf8Vars = [...]string{"LC_ALL", "LC_CTYPE", "LANG"}

// utf8 reports whether the locale promises multi-byte glyphs; nothing set means no.
func utf8(env []string) bool {
	for _, name := range utf8Vars {
		value, ok := lookup(env, name)
		if !ok || value == "" {
			continue
		}
		return isUTF8(value)
	}
	return false
}

// isUTF8 reads the codeset between the dot and any @modifier.
func isUTF8(value string) bool {
	_, codeset, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	codeset, _, _ = strings.Cut(codeset, "@")

	switch strings.ToLower(codeset) {
	case "utf-8", "utf8":
		return true
	default:
		return false
	}
}

func lookup(env []string, name string) (string, bool) {
	prefix := name + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return entry[len(prefix):], true
		}
	}
	return "", false
}
