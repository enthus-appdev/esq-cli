package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Colors for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorCyan   = "\033[0;36m"
	colorDim    = "\033[2m"
)

// colorEnabled checks if stderr is a terminal by checking NO_COLOR env var
// and attempting a simple ioctl-like check via file stat.
var colorEnabled = func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}()

// Info prints an informational message to stderr (so it doesn't interfere with piped JSON).
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", colorCyan, msg, colorReset)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Warn prints a warning message to stderr.
func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", colorYellow, msg, colorReset)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Err prints an error message to stderr.
func Err(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		fmt.Fprintf(os.Stderr, "%serror: %s%s\n", colorRed, msg, colorReset)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
}

// Dim prints dimmed text to stderr.
func Dim(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", colorDim, msg, colorReset)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Success prints a success indicator.
func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", colorGreen, msg, colorReset)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// JSON writes pretty-printed JSON to the given writer.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// RawJSON writes pre-encoded JSON (from ES responses) with consistent formatting.
func RawJSON(w io.Writer, data json.RawMessage) error {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		_, writeErr := w.Write(data)
		return writeErr
	}
	return JSON(w, parsed)
}

// Raw writes a raw string to the given writer.
func Raw(w io.Writer, s string) {
	_, _ = fmt.Fprint(w, s)
}
