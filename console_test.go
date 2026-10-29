// console_test.go

package s99term_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/smegg99/s99logger"
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
	// Step.End measures its own duration, so the clock has to be predictable.
	// Every reading is 1.5s after the one before it.
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return s99term.New(s99term.Options{
		Out:     out,
		Env:     []string{"LANG=C"},
		Words:   words(),
		Profile: profile,
		Animate: &off,
		Now:     func() time.Time { clock = clock.Add(1500 * time.Millisecond); return clock },
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

// A step whose console cannot animate prints only the finished line.
func TestNonAnimatedStepPrintsOnlyTheResult(t *testing.T) {
	c, out := testConsole(t, colorprofile.NoTTY)
	step := c.Begin("building")
	step.End(s99term.StatusOK)

	got := out.String()
	if strings.Contains(got, "\r") || strings.Contains(got, "\x1b") {
		t.Errorf("a non-animated step wrote control bytes: %q", got)
	}
	if want := "[+] ok       building  took=1.5s\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The ticker and End share a mutex. Run under -race.
func TestStepCyclesAreRaceFree(t *testing.T) {
	out := &bytes.Buffer{}
	on := true
	c := s99term.New(s99term.Options{
		Out: out, Env: []string{"LANG=C"}, Words: words(),
		Profile: colorprofile.ASCII, Animate: &on,
	})
	defer c.Close()

	for i := 0; i < 300; i++ {
		step := c.BeginCounted("building", 4)
		time.Sleep(time.Duration(i%20) * time.Millisecond / 10)
		step.Advance(1)
		c.Line(s99term.StatusWarning, "a log record arrived mid-step")
		step.Advance(3)
		step.End(s99term.StatusOK)
	}
}

// A record logged during a step appears above the spinner, and End prints the step's own result.
func TestSinkWritesThroughTheConsole(t *testing.T) {
	c, out := testConsole(t, colorprofile.ASCII)
	log := s99logger.New(c.Sink(), s99logger.Options{Service: "test"})
	log.Warn(s99logger.NewEvent("stale_key", s99logger.String("locale", "cs")))

	got := ansi.Strip(out.String())
	if !strings.HasPrefix(got, "[!] warning") {
		t.Errorf("got %q, want it to start with %q", got, "[!] warning")
	}
	if !strings.Contains(got, "locale=cs") {
		t.Errorf("got %q, want it to carry the attr", got)
	}
}

// The progress bar follows the charset axis, like the spinner.
func TestAsciiCharsetDrawsAnAsciiBar(t *testing.T) {
	out := &bytes.Buffer{}
	on := true
	c := s99term.New(s99term.Options{
		Out: out, Env: []string{"LANG=C"}, Words: words(),
		Profile: colorprofile.ASCII, Animate: &on,
	})
	step := c.BeginCounted("building", 4)
	step.Advance(2)
	step.End(s99term.StatusOK)
	c.Close()

	for i, b := range out.Bytes() {
		if b > 0x7f {
			t.Fatalf("byte %d is %#x; an ASCII charset wrote a multi-byte glyph: %q", i, b, out.String())
		}
	}
}

// A second End on the same step prints nothing.
func TestStepEndsOnce(t *testing.T) {
	c, out := testConsole(t, colorprofile.NoTTY)
	step := c.Begin("building")
	step.End(s99term.StatusOK)
	step.End(s99term.StatusFailed)
	if got := strings.Count(out.String(), "\n"); got != 1 {
		t.Fatalf("End printed %d lines, want 1:\n%s", got, out.String())
	}
}

// A nil Env reads the process environment.
func TestNilEnvReadsTheProcessEnvironment(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_CTYPE", "")
	t.Setenv("LANG", "pl_PL.UTF-8")
	if !s99term.New(s99term.Options{Out: &bytes.Buffer{}}).Charset() {
		t.Fatal("a UTF-8 LANG was not read")
	}
	t.Setenv("LANG", "C")
	if s99term.New(s99term.Options{Out: &bytes.Buffer{}}).Charset() {
		t.Fatal("LANG=C read as UTF-8")
	}
}

// A log line keeps the record's time, service, and whether it was debug.
func TestSinkKeepsTimeServiceAndDebug(t *testing.T) {
	c, out := testConsole(t, colorprofile.NoTTY)
	at := time.Date(2026, 1, 1, 9, 30, 5, 0, time.Local)
	log := s99logger.New(c.Sink(), s99logger.Options{Service: "api", Clock: func() time.Time { return at }})
	log.Debug(s99logger.NewEvent("cache_miss"))
	log.Info(s99logger.NewEvent("ready"))

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines:\n%s", len(lines), out.String())
	}
	for _, want := range []string{"cache_miss", "level=debug", "service=api", "time=09:30:05"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("debug line %q lacks %q", lines[0], want)
		}
	}
	if strings.Contains(lines[1], "level=") {
		t.Errorf("info line %q names its level", lines[1])
	}
}
