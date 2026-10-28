# S99Term

S99Term is the terminal render layer for smegg99 command-line programs. It
detects what the terminal can do on three independent axes and degrades to
readable plain text at the bottom of each one.

Built with lipgloss, bubbles, colorprofile and s99logger.

## What It Does

`Console` owns one output stream and is the only thing in the package that
writes. It takes the stream from `Options.Out`, so nothing here ever reaches
stdout; the consuming CLI passes `os.Stderr` and keeps stdout for machine
output. `Renderer` composes plain strings and holds no stream, so every shape
is testable without a terminal. `Styles` holds every colour, border and pad
value, so no screen builds one inline.

A state is a glyph, a word and a colour, so a line reads the same without
colour. Every glyph in `status.go` is exactly three cells wide and the word
column is padded to a shared width, so the label starts at the same index on
every line whatever the state is.

The package holds no human sentence. `Words` is a `map[Status]string` the
consuming CLI supplies from its own locale catalog. The only literals here are
field keys, which are identifiers and are never translated.

`Step` drives the bubbles spinner and progress components from a plain ticker.
There is no Bubble Tea program, so nothing writes a terminal query and nothing
leaves a reply in the input queue for the shell. `Console.Sink()` returns an
`s99logger.Sink`, so a CLI has one writer for its logs and its presentation,
and a log record that arrives while a step is animating is printed above the
live line rather than through it.

### The three axes

Each axis is detected separately, has its own override in `Options`, and falls
back on its own. None of them costs a terminal round trip.

| Axis | Detected from | Override | Without it |
|---|---|---|---|
| Colour | `colorprofile.Detect(Out, Env)` in `console.go` | `Options.Profile` | `colorprofile.Writer` strips the escapes the stream cannot carry; the glyph and the word stay |
| Animation | whether `Out` is an `*os.File` on a terminal with a width | `Options.Animate` | `Step` prints nothing until `End`, which writes one finished line |
| Charset | `UTF8` in `charset.go`, reading `LC_ALL`, `LC_CTYPE` and `LANG` in that order | `Options.Charset` | `NewStyles` picks `lipgloss.ASCIIBorder`, `Console.frames` picks `spinner.Line`, and a counted step's bar fills with `#` and `-` |

Colour and charset are detected separately. A 16-colour UTF-8 terminal draws
box borders, and a truecolor terminal under `LANG=C` does not, because it cannot
render U+2500.

## Prerequisites

### Build Dependencies

- Go 1.27, the version `go.mod` declares. No cgo.
- `just`, which runs every recipe in `justfile`.
- `git`, optional. Without it `TestTagAgreesWithVersionFile` skips.

This is a library. It builds no binary of its own.

## Build

```sh
just check
```

`check` runs `check-fmt`, `vet`, `test` and `race` in that order. `check-fmt`
fails when a file is not gofmt clean; `fmt` is the separate recipe that
rewrites them. `race` matters here because the step ticker and `Step.End`
share one mutex.

```sh
just demo
```

`demo` runs `internal/demo`, which prints one line of every state, one table
and one counted step, and writes one JSON object to stdout so the two streams
can be told apart.

## Usage

```go
import (
	"os"

	"github.com/smegg99/s99logger"

	"github.com/smegg99/s99term"
)

c := s99term.New(s99term.Options{
	Out:   os.Stderr,
	Env:   os.Environ(),
	Words: s99term.Words{s99term.StatusOK: "ok", s99term.StatusFailed: "failed"},
})
defer c.Close()

log := s99logger.New(c.Sink(), s99logger.Options{Service: "demo"})

step := c.BeginCounted("building locales", 4)
step.Advance(4)
step.End(s99term.StatusOK, s99term.F("keys", "412"))

c.Line(s99term.StatusSkipped, "desktop target not configured")
```

`Options.Words` is the only place human text enters the package. Leave a state
out of the map and its line carries the glyph alone.

This is what `just demo` writes to stderr with `LANG=C` and stderr piped, so
every axis is at its fallback:

```text
s99term 1.0.0
[!] warning  stale_key  locale=cs count=3
[+] ok       building locales  keys=412 took=1.001s
[~] stale    checking drift  locale=cs took=501ms
[x] failed   building cs  reason="missing catalog"
[-] skipped  desktop target not configured
[ ] waiting  queued behind the sync lock
[*] running  watching for changes
             catalog_loaded  locales=3
+-----------+--------+-------+
| KEY       | LOCALE | STATE |
+-----------+--------+-------+
| app.title | pl     | ok    |
| app.sub   | cs     | stale |
+-----------+--------+-------+
```

## Version

`VERSION` holds `MAJOR.MINOR.PATCH` and is the only place the version is
written. `version.go` embeds it. `Declared()` returns exactly that string.
A tagged commit whose
`VERSION` disagrees with its tag fails `TestTagAgreesWithVersionFile`.

## License

MIT. See [LICENSE](LICENSE).
