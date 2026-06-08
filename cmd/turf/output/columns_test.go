package output

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseColumns(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []Column
		wantErr bool
	}{
		{
			name: "single column",
			spec: "NUM:.num",
			want: []Column{
				{Header: "NUM", Path: []string{"num"}},
			},
		},
		{
			name: "multiple columns",
			spec: "NUM:.num,NAME:.name,GRADE:.grade",
			want: []Column{
				{Header: "NUM", Path: []string{"num"}},
				{Header: "NAME", Path: []string{"name"}},
				{Header: "GRADE", Path: []string{"grade"}},
			},
		},
		{
			name: "nested path",
			spec: "COURSE:.fixture.course",
			want: []Column{
				{Header: "COURSE", Path: []string{"fixture", "course"}},
			},
		},
		{
			name: "array expansion",
			spec: "HORSE:.entries[*].horse.name",
			want: []Column{
				{Header: "HORSE", Path: []string{"entries", "[*]", "horse", "name"}},
			},
		},
		{
			name:    "missing colon",
			spec:    "NUM",
			wantErr: true,
		},
		{
			name:    "empty header",
			spec:    ":.num",
			wantErr: true,
		},
		{
			name:    "path without leading dot",
			spec:    "NUM:num",
			wantErr: true,
		},
		{
			name:    "empty spec",
			spec:    "",
			wantErr: true,
		},
		{
			name:    "empty path after dot",
			spec:    "NUM:.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseColumns(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtractRows(t *testing.T) {
	tests := []struct {
		name string
		data any
		cols []Column
		want [][]string
	}{
		{
			name: "flat array",
			data: []map[string]any{
				{"num": 1, "name": "レース1", "grade": "G1"},
				{"num": 2, "name": "レース2", "grade": "G2"},
			},
			cols: []Column{
				{Header: "NUM", Path: []string{"num"}},
				{Header: "NAME", Path: []string{"name"}},
				{Header: "GRADE", Path: []string{"grade"}},
			},
			want: [][]string{
				{"1", "レース1", "G1"},
				{"2", "レース2", "G2"},
			},
		},
		{
			name: "nested object",
			data: []map[string]any{
				{"num": 1, "fixture": map[string]any{"course": "Tokyo"}},
			},
			cols: []Column{
				{Header: "NUM", Path: []string{"num"}},
				{Header: "COURSE", Path: []string{"fixture", "course"}},
			},
			want: [][]string{
				{"1", "Tokyo"},
			},
		},
		{
			name: "array expansion with [*]",
			data: map[string]any{
				"entries": []any{
					map[string]any{"finish": map[string]any{"position": 1, "status": "normal"}, "horse": map[string]any{"name": "ウマA"}},
					map[string]any{"finish": map[string]any{"position": 2, "status": "normal"}, "horse": map[string]any{"name": "ウマB"}},
					map[string]any{"finish": map[string]any{"position": 3, "status": "normal"}, "horse": map[string]any{"name": "ウマC"}},
				},
			},
			cols: []Column{
				{Header: "POS", Path: []string{"entries", "[*]", "finish", "position"}},
				{Header: "HORSE", Path: []string{"entries", "[*]", "horse", "name"}},
			},
			want: [][]string{
				{"1", "ウマA"},
				{"2", "ウマB"},
				{"3", "ウマC"},
			},
		},
		{
			name: "missing field returns empty string",
			data: []map[string]any{
				{"num": 1},
			},
			cols: []Column{
				{Header: "NUM", Path: []string{"num"}},
				{Header: "MISSING", Path: []string{"missing"}},
			},
			want: [][]string{
				{"1", ""},
			},
		},
		{
			name: "single object without array expansion",
			data: map[string]any{
				"course": "Tokyo",
				"year":   2025,
			},
			cols: []Column{
				{Header: "COURSE", Path: []string{"course"}},
				{Header: "YEAR", Path: []string{"year"}},
			},
			want: [][]string{
				{"Tokyo", "2025"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractRows(tt.data, tt.cols)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
