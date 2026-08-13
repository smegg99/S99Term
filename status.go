// status.go

package s99term

// Status is what a line says happened, independent of how it is coloured.
type Status int

// StatusNone is an info-level log record: no state, but aligned under step lines.
const (
	StatusNone Status = iota
	StatusOK
	StatusFailed
	StatusSkipped
	StatusWaiting
	StatusRunning
	StatusWarning
	StatusStale
)

// Glyph is the three-cell marker for st, so colour is never the only carrier.
func (st Status) Glyph() string {
	switch st {
	case StatusOK:
		return "[+]"
	case StatusFailed:
		return "[x]"
	case StatusSkipped:
		return "[-]"
	case StatusWaiting:
		return "[ ]"
	case StatusRunning:
		return "[*]"
	case StatusWarning:
		return "[!]"
	case StatusStale:
		return "[~]"
	default:
		return "   "
	}
}

// Words maps each state to its localized name, supplied by the consuming CLI.
type Words map[Status]string
