package jra

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCommaSeparatedIntUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    commaSeparatedInt
		wantErr bool
	}{
		{"1600", "1600", commaSeparatedInt(1600), false},
		{"2,400", "2,400", commaSeparatedInt(2400), false},
		{"invalid", "abc", commaSeparatedInt(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got commaSeparatedInt
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("distance mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

