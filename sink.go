// sink.go

package s99term

import (
	"context"
	"fmt"

	"github.com/smegg99/s99logger"
)

// Sink returns an s99logger.Sink that renders records through the shared
// styles. The logger keeps owning the event stream; this owns its presentation.
func (c *Console) Sink() s99logger.Sink { return sink{c: c} }

type sink struct{ c *Console }

// Write renders rec as one status line, above a live step's frame if there is
// one.
func (s sink) Write(_ context.Context, rec s99logger.Record) error {
	fields := make([]Field, 0, len(rec.Attrs)+1)
	for _, a := range rec.Attrs {
		fields = append(fields, F(a.Key, fmt.Sprintf("%v", a.Value)))
	}
	if rec.Message != string(rec.MessageID) {
		fields = append(fields, F("id", string(rec.MessageID)))
	}
	s.c.Line(statusForLevel(rec.Level), rec.Message, fields...)
	return nil
}

// statusForLevel maps a log level onto the shared state vocabulary.
func statusForLevel(level s99logger.Level) Status {
	switch level {
	case s99logger.LevelError:
		return StatusFailed
	case s99logger.LevelWarn:
		return StatusWarning
	default:
		return StatusNone
	}
}
