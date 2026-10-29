// console.go

package s99term

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

// Options configures a Console.
type Options struct {
	// Out receives every line. Pass os.Stderr to keep stdout for machine output.
	Out io.Writer
	// Env feeds colour and charset detection. Nil reads os.Environ().
	Env []string
	// Words names each state. A state left out prints its glyph alone.
	Words Words
	// Profile overrides colour detection. The zero value detects.
	Profile colorprofile.Profile
	// Animate overrides animation. Nil animates only on a sized terminal.
	Animate *bool
	// Charset overrides the charset answer. Nil reads the locale variables.
	Charset *bool
	// Now supplies step timestamps. Nil uses time.Now.
	Now func() time.Time
}

// Console owns one stream: its colour level, charset, width and the one live step on it.
type Console struct {
	mu       sync.Mutex
	out      io.Writer
	w        *colorprofile.Writer
	renderer Renderer
	profile  colorprofile.Profile
	animate  bool
	charset  bool
	width    int
	step     *Step
	now      func() time.Time
}

// New resolves the colour, animation and charset axes and builds the writer.
func New(opts Options) *Console {
	// Colour comes from the environment, animation from the file descriptor,
	// and charset from the locale variables. None costs a terminal round trip.
	out := opts.Out
	if out == nil {
		out = os.Stderr
	}
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}

	// Unknown is colorprofile's zero value, so an unset Options.Profile means
	// detect and every other value is an explicit override from --color.
	profile := opts.Profile
	if profile == colorprofile.Unknown {
		profile = colorprofile.Detect(out, env)
	}

	animate, width := false, 80
	if file, ok := out.(*os.File); ok {
		fd := file.Fd()
		if term.IsTerminal(fd) {
			if w, _, err := term.GetSize(fd); err == nil && w > 0 {
				animate, width = true, w
			}
		}
	}
	if opts.Animate != nil {
		animate = *opts.Animate
	}

	charset := utf8(env)
	if opts.Charset != nil {
		charset = *opts.Charset
	}

	styles := newStyles(charset)
	return &Console{
		// Build the writer from the resolved profile, not from env:
		// colorprofile.NewWriter re-detects and would discard an override.
		w:        &colorprofile.Writer{Forward: out, Profile: profile},
		out:      out,
		profile:  profile,
		renderer: NewRenderer(styles, opts.Words),
		animate:  animate,
		charset:  charset,
		width:    width,
		now:      now,
	}
}

func (c *Console) Renderer() Renderer { return c.renderer }

func (c *Console) Styles() Styles { return c.renderer.Styles }

// Width returns the terminal width, or 80 when the stream has none.
func (c *Console) Width() int { return c.width }

func (c *Console) Charset() bool { return c.charset }

func (c *Console) Profile() colorprofile.Profile { return c.profile }

// Print writes one rendered block, erasing and redrawing a live step's frame around it.
func (c *Console) Print(block string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.printLocked(block)
}

func (c *Console) printLocked(block string) {
	if c.step != nil {
		_, _ = io.WriteString(c.out, "\r"+ansi.EraseLineRight)
	}
	if !strings.HasSuffix(block, "\n") {
		block += "\n"
	}
	_, _ = c.w.WriteString(block)
	if c.step != nil {
		c.step.drawLocked()
	}
}

// Line writes one status line.
func (c *Console) Line(st Status, label string, fields ...Field) {
	c.Print(c.renderer.Line(st, label, fields...))
}

// Close ends any live step and restores the cursor. Defer it in every command.
func (c *Console) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.step != nil {
		c.step.stopLocked()
	}
}
