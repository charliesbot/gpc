package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter renders data as pretty-printed JSON with 2-space indentation.
type JSONFormatter struct {
	Writer io.Writer
}

// Format marshals data to indented JSON and writes it to the underlying
// writer followed by a trailing newline.
func (f *JSONFormatter) Format(data any) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}
