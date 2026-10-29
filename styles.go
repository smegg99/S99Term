// styles.go

package s99term

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/charmbracelet/x/ansi"
)

// accent is the only fixed colour; the rest follow the terminal theme.
const accent = "#c73378"

// Palette holds every colour. All but Accent are ANSI indices, so no background
// query is needed.
type Palette struct {
	Accent color.Color
	Muted  color.Color
	Line   color.Color
	Good   color.Color
	Bad    color.Color
	Warn   color.Color
}

// Styles owns every colour, border and pad value in the package.
type Styles struct {
	Palette     Palette
	Border      lipgloss.Border
	Unicode     bool
	Title       lipgloss.Style
	Ink         lipgloss.Style
	Muted       lipgloss.Style
	AccentStyle lipgloss.Style
	Line        lipgloss.Style
	Good        lipgloss.Style
	Bad         lipgloss.Style
	Warn        lipgloss.Style
	Key         lipgloss.Style
	Field       lipgloss.Style
	Spinner     lipgloss.Style
	Panel       lipgloss.Style
	StatusWidth int
}

// newStyles builds the palette. unicode is the charset answer from UTF8, not
// the colour profile: a 16-colour UTF-8 terminal still draws box borders.
func newStyles(unicode bool) Styles {
	palette := Palette{
		Accent: lipgloss.Color(accent),
		Muted:  lipgloss.Color("8"),
		Line:   lipgloss.Color("8"),
		Good:   lipgloss.Color("2"),
		Bad:    lipgloss.Color("1"),
		Warn:   lipgloss.Color("3"),
	}

	border := lipgloss.ASCIIBorder()
	if unicode {
		border = lipgloss.NormalBorder()
	}

	return Styles{
		Palette:     palette,
		Border:      border,
		Unicode:     unicode,
		Title:       lipgloss.NewStyle().Foreground(palette.Accent).Bold(true),
		Ink:         lipgloss.NewStyle(),
		Muted:       lipgloss.NewStyle().Foreground(palette.Muted),
		AccentStyle: lipgloss.NewStyle().Foreground(palette.Accent),
		Line:        lipgloss.NewStyle().Foreground(palette.Line),
		Good:        lipgloss.NewStyle().Foreground(palette.Good),
		Bad:         lipgloss.NewStyle().Foreground(palette.Bad),
		Warn:        lipgloss.NewStyle().Foreground(palette.Warn),
		Key:         lipgloss.NewStyle().Foreground(palette.Muted),
		Field:       lipgloss.NewStyle(),
		Spinner:     lipgloss.NewStyle().Foreground(palette.Accent),
		Panel:       lipgloss.NewStyle().Padding(0, 1),
		StatusWidth: 12,
	}
}

// statusStyle is the colour for st. The glyph and the word carry the meaning on
// their own, so this only reinforces them.
func (s Styles) statusStyle(st Status) lipgloss.Style {
	switch st {
	case StatusOK:
		return s.Good
	case StatusFailed:
		return s.Bad
	case StatusWarning, StatusStale:
		return s.Warn
	case StatusRunning:
		return s.AccentStyle
	case StatusSkipped, StatusWaiting:
		return s.Muted
	default:
		return s.Ink
	}
}

// Table is the one table shape in the package.
func (s Styles) Table() *table.Table {
	return table.New().
		Border(s.Border).
		BorderStyle(s.Line).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return s.Key.Padding(0, 1)
			}
			return s.Ink.Padding(0, 1)
		})
}

// statusWidth is the column the label starts in: the glyph, a space, the widest
// word and a space, with a floor so a short catalog still lines up.
func statusWidth(words Words) int {
	widest := 0
	for _, word := range words {
		if w := ansi.StringWidth(word); w > widest {
			widest = w
		}
	}
	if width := 3 + 1 + widest + 1; width > 12 {
		return width
	}
	return 12
}
