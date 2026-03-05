package output

import (
	"fmt"
	"os"
	"strings"
)

const barWidth = 30

// ProgressBar provides a simple terminal progress bar that prints to stderr.
// It implements io.Writer so it can wrap upload streams to report progress
// automatically.
type ProgressBar struct {
	total       int64
	current     int64
	description string
}

// NewProgressBar creates a ProgressBar for an operation of the given total
// size. The description is displayed alongside the bar (e.g. "Uploading...").
func NewProgressBar(total int64, description string) *ProgressBar {
	return &ProgressBar{
		total:       total,
		description: description,
	}
}

// Write implements io.Writer. Each call advances the progress bar by
// len(b) bytes and re-renders the bar on stderr using a carriage return
// so it updates in place.
func (p *ProgressBar) Write(b []byte) (int, error) {
	n := len(b)
	p.current += int64(n)
	p.render()
	return n, nil
}

// Finish marks the progress as complete, renders the final state, and
// prints a newline so subsequent output starts on a fresh line.
func (p *ProgressBar) Finish() {
	p.current = p.total
	p.render()
	fmt.Fprintln(os.Stderr)
}

// render draws the progress bar to stderr. The bar looks like:
//
//	[████████████░░░░░░░░░░░░░░░░░░] 40% Uploading...
func (p *ProgressBar) render() {
	var pct float64
	if p.total > 0 {
		pct = float64(p.current) / float64(p.total)
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(pct * float64(barWidth))
	empty := barWidth - filled

	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", empty)
	fmt.Fprintf(os.Stderr, "\r[%s] %3.0f%% %s", bar, pct*100, p.description)
}
