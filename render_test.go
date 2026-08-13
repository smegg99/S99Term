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
