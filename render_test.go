// render_test.go

package s99term

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Every glyph is three cells, so the label column starts at the same index on
// every line whatever the state is.
func TestEveryGlyphIsThreeCells(t *testing.T) {
	for _, st := range []Status{
		StatusOK, StatusFailed, StatusSkipped,
		StatusWaiting, StatusRunning, StatusWarning,
		StatusStale, StatusNone,
	} {
		if got := ansi.StringWidth(st.Glyph()); got != 3 {
			t.Errorf("Status(%d).Glyph() = %q, width %d, want 3", st, st.Glyph(), got)
		}
	}
}

func TestGlyphsAreDistinctExceptNone(t *testing.T) {
	seen := map[string]Status{}
	for _, st := range []Status{
		StatusOK, StatusFailed, StatusSkipped,
		StatusWaiting, StatusRunning, StatusWarning,
		StatusStale,
	} {
		if prev, dup := seen[st.Glyph()]; dup {
			t.Errorf("Status(%d) and Status(%d) share glyph %q", prev, st, st.Glyph())
		}
		seen[st.Glyph()] = st
	}
}

func TestUnicodeDecidesTheBorderAndNotTheColour(t *testing.T) {
	unicode := newStyles(true)
	ascii := newStyles(false)

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
	english := Words{StatusOK: "ok", StatusWarning: "warning"}
	polish := Words{StatusOK: "ok", StatusWarning: "ostrzezenie"}

	if got := statusWidth(english); got != 12 {
		t.Errorf("StatusWidth(english) = %d, want the 12 floor", got)
	}
	if got, want := statusWidth(polish), 3+1+11+1; got != want {
		t.Errorf("StatusWidth(polish) = %d, want %d", got, want)
	}
}

func TestLineIsGlyphWordAndLabelInFixedColumns(t *testing.T) {
	r := NewRenderer(newStyles(false), Words{
		StatusOK:      "ok",
		StatusFailed:  "failed",
		StatusWarning: "warning",
	})

	ok := ansi.Strip(r.Line(StatusOK, "built pl", F("keys", "412")))
	failed := ansi.Strip(r.Line(StatusFailed, "built pl"))

	if want := "[+] ok       built pl  keys=412"; ok != want {
		t.Errorf("Line ok = %q, want %q", ok, want)
	}
	if got, want := strings.Index(ok, "built"), strings.Index(failed, "built"); got != want {
		t.Errorf("label starts at %d for ok and %d for failed; columns must agree", got, want)
	}
}

// A value with a newline would break the one-line redraw, which clears one line.
func TestFieldsQuoteWhitespaceAndQuotes(t *testing.T) {
	r := NewRenderer(newStyles(false), Words{})

	for _, c := range []struct{ in, want string }{
		{"412", "n=412"},
		{"two words", `n="two words"`},
		{"line\nbreak", `n="line\nbreak"`},
		{`say "hi"`, `n="say \"hi\""`},
	} {
		if got := ansi.Strip(r.Fields(F("n", c.in))); got != c.want {
			t.Errorf("Fields(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
