package output

import (
	"fmt"
	"io"
	"strings"
)

// Write writes data to w in the specified format.
// Supported formats: "json", "custom-columns=HEADER:.path,...", or "" (default table).
func Write(w io.Writer, data any, format string, defaultColumns []Column) error {
	if format == "json" {
		return WriteJSON(w, data)
	}

	if strings.HasPrefix(format, "custom-columns=") {
		spec := strings.TrimPrefix(format, "custom-columns=")
		cols, err := ParseColumns(spec)
		if err != nil {
			return fmt.Errorf("invalid custom-columns: %w", err)
		}
		return writeTable(w, data, cols)
	}

	if format == "" {
		return writeTable(w, data, defaultColumns)
	}

	return fmt.Errorf("unknown output format %q: supported formats are json, custom-columns=SPEC", format)
}

func writeTable(w io.Writer, data any, cols []Column) error {
	rows, err := ExtractRows(data, cols)
	if err != nil {
		return err
	}

	headers := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = col.Header
	}

	return WriteTable(w, headers, rows)
}
