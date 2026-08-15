// render_test.go

package s99term_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/smegg99/s99term"
)

// Every glyph is three cells, so the label column starts at the same index on
// every line whatever the state is.
func TestEveryGlyphIsThreeCells(t *testing.T) {
	for _, st := range []s99term.Status{
		s99term.StatusOK, s99term.StatusFailed, s99term.StatusSkipped,
		s99term.StatusWaiting, s99term.StatusRunning, s99term.StatusWarning,
		s99term.StatusStale, s99term.StatusNone,
	} {
		if got := ansi.StringWidth(st.Glyph()); got != 3 {
			t.Errorf("Status(%d).Glyph() = %q, width %d, want 3", st, st.Glyph(), got)
		}
	}
}

func TestGlyphsAreDistinctExceptNone(t *testing.T) {
	seen := map[string]s99term.Status{}
	for _, st := range []s99term.Status{
		s99term.StatusOK, s99term.StatusFailed, s99term.StatusSkipped,
		s99term.StatusWaiting, s99term.StatusRunning, s99term.StatusWarning,
		s99term.StatusStale,
	} {
		if prev, dup := seen[st.Glyph()]; dup {
			t.Errorf("Status(%d) and Status(%d) share glyph %q", prev, st, st.Glyph())
		}
		seen[st.Glyph()] = st
	}
}

func TestUnicodeDecidesTheBorderAndNotTheColour(t *testing.T) {
	unicode := s99term.NewStyles(true)
	ascii := s99term.NewStyles(false)

	if unicode.Border.Top == ascii.Border.Top {
		t.Fatal("unicode and ascii styles share a border")
	}
	if ascii.Border.Top != "-" {
		t.Errorf("ascii border top = %q, want %q", ascii.Border.Top, "-")
	}
	if unicode.Palette.Good != ascii.Palette.Good {
		t.Error("the charset axis changed a colour, which is the colour axis")
	}
}

// A Polish catalog has longer words than an English one, so the column widens
// rather than pushing the label out of line.
func TestStatusWidthFollowsTheWidestWord(t *testing.T) {
	english := s99term.Words{s99term.StatusOK: "ok", s99term.StatusWarning: "warning"}
	polish := s99term.Words{s99term.StatusOK: "ok", s99term.StatusWarning: "ostrzezenie"}

	if got := s99term.StatusWidth(english); got != 12 {
		t.Errorf("StatusWidth(english) = %d, want the 12 floor", got)
	}
	if got, want := s99term.StatusWidth(polish), 3+1+11+1; got != want {
		t.Errorf("StatusWidth(polish) = %d, want %d", got, want)
	}
}
