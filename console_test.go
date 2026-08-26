// console_test.go

package s99term_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/smegg99/s99term"
)

func words() s99term.Words {
	return s99term.Words{
		s99term.StatusOK:      "ok",
		s99term.StatusFailed:  "failed",
		s99term.StatusSkipped: "skipped",
		s99term.StatusWaiting: "waiting",
		s99term.StatusRunning: "running",
		s99term.StatusWarning: "warning",
		s99term.StatusStale:   "stale",
	}
}

func testConsole(t *testing.T, profile colorprofile.Profile) (*s99term.Console, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	off := false
	return s99term.New(s99term.Options{
		Out:     out,
		Env:     []string{"LANG=C"},
		Words:   words(),
		Profile: profile,
		Animate: &off,
	}), out
}

// Without colour a line still carries its glyph and word.
func TestAsciiProfileStripsColourAndKeepsMeaning(t *testing.T) {
	c, out := testConsole(t, colorprofile.ASCII)
	c.Line(s99term.StatusOK, "built pl", s99term.F("keys", "412"))

	got := out.String()
	if strings.Contains(got, "\x1b[3") {
		t.Errorf("colour survived the Ascii profile: %q", got)
	}
	if want := "[+] ok       built pl  keys=412\n"; !strings.Contains(ansi.Strip(got), want) {
		t.Errorf("got %q, want it to contain %q", got, want)
	}
}

// Every rung of the ladder carries the same glyph and word.
func TestEveryProfileKeepsGlyphAndWord(t *testing.T) {
	for _, profile := range []colorprofile.Profile{
		colorprofile.TrueColor, colorprofile.ANSI256,
		colorprofile.ANSI, colorprofile.ASCII, colorprofile.NoTTY,
	} {
		c, out := testConsole(t, profile)
		c.Line(s99term.StatusFailed, "build pl")

		if got := ansi.Strip(out.String()); !strings.HasPrefix(got, "[x] failed") {
			t.Errorf("profile %v produced %q, want it to start with %q", profile, got, "[x] failed")
		}
	}
}

func TestNoTTYStripsEveryEscape(t *testing.T) {
	c, out := testConsole(t, colorprofile.NoTTY)
	c.Line(s99term.StatusWarning, "stale key", s99term.F("locale", "cs"))

	if got := out.String(); strings.Contains(got, "\x1b") {
		t.Errorf("an escape survived NoTTY: %q", got)
	}
}
