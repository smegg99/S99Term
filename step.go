// step.go

package s99term

import (
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"

	"github.com/charmbracelet/x/ansi"
)

// Step is one unit of work the reader waits on. On a terminal it animates one
// inline line; on any other stream it prints nothing until it ends.
type Step struct {
	c       *Console
	label   string
	started time.Time
	sp      spinner.Model
	bar     progress.Model
	total   int
	done    int
	stop    chan struct{}
}

// Begin starts a step whose end is not countable.
func (c *Console) Begin(label string) *Step { return c.begin(label, 0) }

// BeginCounted starts a step with a known number of units, shown as a bar.
func (c *Console) BeginCounted(label string, total int) *Step {
	return c.begin(label, total)
}

// frames picks the spinner the terminal can actually draw. MiniDot is braille,
// which an ASCII terminal renders as replacement characters.
func (c *Console) frames() spinner.Spinner {
	if c.charset {
		return spinner.MiniDot
	}
	return spinner.Line
}

func (c *Console) begin(label string, total int) *Step {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.step != nil {
		c.step.stopLocked()
	}
	s := &Step{c: c, label: label, started: c.now(), total: total}
	if !c.animate {
		return s
	}
	s.sp = spinner.New(spinner.WithSpinner(c.frames()))
	if total > 0 {
		s.bar = progress.New(
			progress.WithColors(c.renderer.Styles.Palette.Accent),
			progress.WithoutPercentage(),
		)
		s.bar.EmptyColor = c.renderer.Styles.Palette.Line
		// The bar follows the charset axis for the same reason the spinner
		// does: its default fill is U+258C and U+2591.
		if !c.charset {
			s.bar.Full, s.bar.Empty = '#', '-'
		}
		s.bar.SetWidth(barWidth(c.width))
	}
	s.stop = make(chan struct{})
	c.step = s
	_, _ = io.WriteString(c.out, ansi.HideCursor)
	s.drawLocked()
	go s.run()
	return s
}

func barWidth(width int) int {
	w := width / 4
	if w < 10 {
		return 10
	}
	if w > 30 {
		return 30
	}
	return w
}

// run advances the spinner until the step ends. A late tick finds c.step != s
// and repaints nothing, so End never waits on run.
func (s *Step) run() {
	tick := time.NewTicker(s.sp.Spinner.FPS)
	defer tick.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-tick.C:
			s.c.mu.Lock()
			if s.c.step == s {
				s.sp, _ = s.sp.Update(s.sp.Tick())
				s.drawLocked()
			}
			s.c.mu.Unlock()
		}
	}
}

// Advance adds n finished units to a counted step.
func (s *Step) Advance(n int) {
	s.c.mu.Lock()
	defer s.c.mu.Unlock()
	s.done += n
	if s.c.step == s {
		s.drawLocked()
	}
}

// drawLocked repaints the live line. The console mutex must be held.
func (s *Step) drawLocked() {
	if s.c.step != s {
		return
	}
	trailer := ""
	if s.total > 0 {
		trailer = s.bar.ViewAs(s.percent()) + " " +
			s.c.renderer.Styles.Muted.Render(fmt.Sprintf("%d/%d", s.done, s.total))
	}
	frame := s.c.renderer.Frame(s.sp.View(), s.label, trailer)
	frame = ansi.Truncate(frame, s.c.width-1, "")
	_, _ = io.WriteString(s.c.out, "\r"+ansi.EraseLineRight)
	_, _ = s.c.w.WriteString(frame)
}

func (s *Step) percent() float64 {
	if s.total <= 0 {
		return 0
	}
	return float64(s.done) / float64(s.total)
}

// stopLocked ends the animation and clears the line. The mutex must be held.
func (s *Step) stopLocked() {
	if s.c.step != s {
		return
	}
	s.c.step = nil
	close(s.stop)
	_, _ = io.WriteString(s.c.out, "\r"+ansi.EraseLineRight+ansi.ShowCursor)
}

// End closes the step and leaves exactly one finished line behind it.
func (s *Step) End(st Status, fields ...Field) {
	s.c.mu.Lock()
	defer s.c.mu.Unlock()
	if s.c.step == s {
		s.stopLocked()
	}
	fields = append(fields, Dur("took", s.c.now().Sub(s.started)))
	s.c.printLocked(s.c.renderer.Line(st, s.label, fields...))
}
