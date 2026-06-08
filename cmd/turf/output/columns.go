package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Column defines a single column in a table output.
// Path is a dot-separated list of JSON field names to traverse,
// with "[*]" as a special token that expands an array into multiple rows.
// Format is an optional function to transform the raw string value before display.
type Column struct {
	Header string
	Path   []string
	Format func(string) string
}

func ParseColumns(spec string) ([]Column, error) {
	var cols []Column
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, ":")
		if idx < 0 {
			return nil, fmt.Errorf("invalid column spec %q: expected HEADER:.path", part)
		}
		header := part[:idx]
		path := part[idx+1:]
		if header == "" {
			return nil, fmt.Errorf("invalid column spec %q: header is empty", part)
		}
		segments, err := parsePath(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q: %w", path, err)
		}
		cols = append(cols, Column{Header: header, Path: segments})
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("no columns specified")
	}
	return cols, nil
}

func parsePath(path string) ([]string, error) {
	if !strings.HasPrefix(path, ".") {
		return nil, fmt.Errorf("path must start with '.'")
	}
	path = path[1:]
	if path == "" {
		return nil, fmt.Errorf("path is empty after '.'")
	}

	var segments []string
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			return nil, fmt.Errorf("empty path segment")
		}
		if strings.HasSuffix(seg, "[*]") {
			field := seg[:len(seg)-3]
			if field != "" {
				segments = append(segments, field)
			}
			segments = append(segments, "[*]")
		} else {
			segments = append(segments, seg)
		}
	}
	return segments, nil
}

// ExtractRows converts structured data into a 2D string table based on the given column definitions.
// data is marshaled to JSON and then traversed using each column's Path.
// Columns with a "[*]" segment cause array expansion, producing multiple rows per item.
func ExtractRows(data any, cols []Column) ([][]string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	topItems := toSlice(v)

	var rows [][]string
	for _, item := range topItems {
		itemRows := extractFromItem(item, cols)
		rows = append(rows, itemRows...)
	}
	return rows, nil
}

func extractFromItem(item any, cols []Column) [][]string {
	colValues := make([][]string, len(cols))
	maxLen := 1
	for i, col := range cols {
		vals := extractValues(item, col.Path)
		colValues[i] = vals
		if len(vals) > maxLen {
			maxLen = len(vals)
		}
	}

	rows := make([][]string, maxLen)
	for r := range rows {
		row := make([]string, len(cols))
		for c := range cols {
			if r < len(colValues[c]) {
				s := colValues[c][r]
				if cols[c].Format != nil {
					s = cols[c].Format(s)
				}
				row[c] = s
			}
		}
		rows[r] = row
	}
	return rows
}

// extractValues recursively traverses a JSON-like value following the given path segments.
// When the segment is "[*]", it iterates all elements of the array and flattens the results.
// Returns a slice of string values; an empty string is returned for missing or null fields.
func extractValues(v any, path []string) []string {
	if len(path) == 0 {
		return []string{stringify(v)}
	}

	seg := path[0]
	rest := path[1:]

	if seg == "[*]" {
		items := toSlice(v)
		var results []string
		for _, item := range items {
			results = append(results, extractValues(item, rest)...)
		}
		return results
	}

	m, ok := v.(map[string]any)
	if !ok {
		return []string{""}
	}
	val, exists := m[seg]
	if !exists {
		return []string{""}
	}
	return extractValues(val, rest)
}

func toSlice(v any) []any {
	if arr, ok := v.([]any); ok {
		return arr
	}
	return []any{v}
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
