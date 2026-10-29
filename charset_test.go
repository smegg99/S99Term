// charset_test.go

package s99term

import (
	"testing"
)

func TestUTF8ReadsThePosixLocalePrecedence(t *testing.T) {
	for _, c := range []struct {
		name string
		env  []string
		want bool
	}{
		{"nothing set", nil, false},
		{"LANG utf-8", []string{"LANG=pl_PL.UTF-8"}, true},
		{"LANG lowercase utf8", []string{"LANG=en_US.utf8"}, true},
		{"LANG C", []string{"LANG=C"}, false},
		{"LANG POSIX", []string{"LANG=POSIX"}, false},
		{"LANG C.UTF-8", []string{"LANG=C.UTF-8"}, true},
		{"LANG empty falls through to nothing", []string{"LANG="}, false},
		{"LC_CTYPE beats LANG", []string{"LANG=C", "LC_CTYPE=pl_PL.UTF-8"}, true},
		{"LC_ALL beats LC_CTYPE", []string{"LC_ALL=C", "LC_CTYPE=pl_PL.UTF-8"}, false},
		{"LC_ALL utf-8 beats a C LANG", []string{"LC_ALL=de_DE.UTF-8", "LANG=C"}, true},
		{"empty LC_ALL falls through to LANG", []string{"LC_ALL=", "LANG=pl_PL.UTF-8"}, true},
		{"modifier after codeset", []string{"LANG=sr_RS.UTF-8@latin"}, true},
		{"no codeset", []string{"LANG=pl_PL"}, false},
		{"unrelated variable ignored", []string{"LANGUAGE=pl_PL.UTF-8"}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := utf8(c.env); got != c.want {
				t.Errorf("UTF8(%v) = %v, want %v", c.env, got, c.want)
			}
		})
	}
}
