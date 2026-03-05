package output

import "io"

// Formatter defines a common interface for rendering structured data to
// a writer. Implementations handle the specifics of each output format
// (e.g. JSON, table).
type Formatter interface {
	Format(data any) error
}

// New creates a Formatter for the given format string. Supported formats
// are "json" and "table". Any unrecognised format defaults to "table".
func New(format string, w io.Writer) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{Writer: w}
	default:
		return &TableFormatter{Writer: w}
	}
}
