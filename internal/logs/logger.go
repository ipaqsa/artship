package logs

import (
	"fmt"
	"io"
	"os"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	verbose bool
	output  io.Writer
}

// New creates a new enhanced logger
func New(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
		output:  os.Stdout,
	}
}

// Info logs a step message (always visible)
func (l *Logger) Info(msg string, args ...any) {
	_, _ = fmt.Fprintf(l.output, "%s\n", fmt.Sprintf(msg, args...))
}

// Debug logs detailed step information (only in verbose mode)
func (l *Logger) Debug(msg string, args ...any) {
	if l.verbose {
		_, _ = fmt.Fprintf(l.output, "[DEBUG] %s\n", fmt.Sprintf(msg, args...))
	}
}
