// internal/demo/main.go

// Command demo prints one of every rendered shape for checking the palette.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/smegg99/s99logger"

	"github.com/smegg99/s99term"
)

func main() {
	c := s99term.New(s99term.Options{
		Out:   os.Stderr,
		Env:   os.Environ(),
		Words: words(),
	})
	defer c.Close()

	log := s99logger.New(c.Sink(), s99logger.Options{Service: "demo"})

	c.Print(c.Renderer().Title("s99term " + s99term.Version()))

	build := c.BeginCounted("building locales", 4)
	for i := 1; i <= 4; i++ {
		time.Sleep(250 * time.Millisecond)
		if i == 2 {
			log.Warn(s99logger.NewEvent("stale_key",
				s99logger.String("locale", "cs"), s99logger.Int("count", 3)))
		}
		build.Advance(1)
	}
	build.End(s99term.StatusOK, s99term.F("keys", "412"))

	drift := c.Begin("checking drift")
	time.Sleep(500 * time.Millisecond)
	drift.End(s99term.StatusStale, s99term.F("locale", "cs"))

	// The states no step produced, so every one of the eight is on screen.
	c.Line(s99term.StatusFailed, "building cs", s99term.F("reason", "missing catalog"))
	c.Line(s99term.StatusSkipped, "desktop target not configured")
	c.Line(s99term.StatusWaiting, "queued behind the sync lock")
	c.Line(s99term.StatusRunning, "watching for changes")
	log.Info(s99logger.NewEvent("catalog_loaded", s99logger.Int("locales", 3)))

	c.Print(c.Styles().Table().
		Headers("KEY", "LOCALE", "STATE").
		Row("app.title", "pl", "ok").
		Row("app.sub", "cs", "stale").
		Render())

	// Stdout stays machine readable while every human byte goes to stderr.
	fmt.Fprintln(os.Stdout, `{"machine":"readable"}`)
}

// words stands in for the catalog a consuming CLI supplies.
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
