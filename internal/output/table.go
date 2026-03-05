package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TableData is the structured input for table rendering, combining column
// headers with row data.
type TableData struct {
	Headers []string
	Rows    [][]string
}

// TableFormatter renders data as an aligned table.
type TableFormatter struct {
	Writer io.Writer
}

// Format writes data as a table to the underlying writer. It accepts either
// a [][]string (headerless rows) or a TableData value/pointer containing
// both headers and rows. Any other type results in an error.
func (f *TableFormatter) Format(data any) error {
	switch v := data.(type) {
	case [][]string:
		return f.render(nil, v)
	case TableData:
		return f.render(v.Headers, v.Rows)
	case *TableData:
		if v == nil {
			return fmt.Errorf("table data is nil")
		}
		return f.render(v.Headers, v.Rows)
	default:
		return fmt.Errorf("unsupported table data type %T; expected [][]string or TableData", data)
	}
}

// render builds and flushes the table with the optional headers and rows.
func (f *TableFormatter) render(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)

	if len(headers) > 0 {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
		dashes := make([]string, len(headers))
		for i, h := range headers {
			dashes[i] = strings.Repeat("-", len(h))
		}
		fmt.Fprintln(w, strings.Join(dashes, "\t"))
	}

	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}
