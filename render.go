// render.go

package s99term

import (
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// Field is one key and value appended to a line. The key is an identifier, like
// an s99logger attr key, so it is never translated.
type Field struct {
	Key   string
	Value string
}

// F is a field with a value that is already a string.
func F(key, value string) Field { return Field{Key: key, Value: value} }

// Dur is a field holding a duration rounded to the millisecond.
func Dur(key string, d time.Duration) Field {
	return Field{Key: key, Value: d.Round(time.Millisecond).String()}
}

// Renderer composes the strings. It holds no stream, so every function here is
// testable without a terminal.
type Renderer struct {
	Styles Styles
	Words  Words
}

// NewRenderer pairs a palette with the catalog words for each state.
func NewRenderer(styles Styles, words Words) Renderer {
	styles.StatusWidth = StatusWidth(words)
	return Renderer{Styles: styles, Words: words}
}

// StatusCell is the glyph and the word, padded to the shared column width.
func (r Renderer) StatusCell(st Status) string {
	style := r.Styles.StatusStyle(st)
	cell := lipgloss.JoinHorizontal(lipgloss.Top,
		style.Render(st.Glyph()),
		" ",
		style.Render(r.Words[st]),
	)
	return lipgloss.PlaceHorizontal(r.Styles.StatusWidth, lipgloss.Left, cell)
}

// Fields renders key=value pairs, quoting any value that would break a line.
func (r Renderer) Fields(fields ...Field) string {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		value := f.Value
		if strings.ContainsAny(value, " \t\n\r\"") {
			value = strconv.Quote(value)
		}
		parts = append(parts, r.Styles.Key.Render(f.Key+"=")+r.Styles.Field.Render(value))
	}
	return strings.Join(parts, " ")
}

// Line is one finished statement: state, label, then whatever it measured.
func (r Renderer) Line(st Status, label string, fields ...Field) string {
	line := lipgloss.JoinHorizontal(lipgloss.Top, r.StatusCell(st), " ", r.Styles.Ink.Render(label))
	if len(fields) == 0 {
		return line
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, line, "  ", r.Fields(fields...))
}

// Title is a heading above a group of lines.
func (r Renderer) Title(text string) string { return r.Styles.Title.Render(text) }

// Frame is the live line while a step runs: spinner, label, and whatever the
// step wants to show on the right, such as a progress bar.
func (r Renderer) Frame(spinner, label, trailer string) string {
	cell := lipgloss.PlaceHorizontal(r.Styles.StatusWidth, lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, r.Styles.Spinner.Render(spinner)))
	line := lipgloss.JoinHorizontal(lipgloss.Top, cell, " ", r.Styles.Ink.Render(label))
	if trailer == "" {
		return line
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, line, "  ", trailer)
}
