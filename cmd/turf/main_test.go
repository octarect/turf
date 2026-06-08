package main

import (
	"testing"
	"time"

	"github.com/octarect/turf/model"
)

func TestParseMonth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{name: "valid", input: "2024-05", want: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)},
		{name: "before earliest", input: "1985-12", wantErr: true},
		{name: "invalid format", input: "2024/05", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMonth(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if !got.Equal(tt.want) {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{name: "valid", input: "2024-05-31", want: time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC)},
		{name: "before earliest", input: "1985-12-31", wantErr: true},
		{name: "invalid format", input: "2024/05/31", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if !got.Equal(tt.want) {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestParseCourse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.Course
		wantErr bool
	}{
		{name: "empty", input: "", want: model.CourseUnknown},
		{name: "valid", input: "tokyo", want: model.CourseTokyo},
		{name: "title case", input: "Tokyo", want: model.CourseTokyo},
		{name: "mixed case", input: "HaNsHiN", want: model.CourseHanshin},
		{name: "camel case", input: "NakaYama", want: model.CourseNakayama},
		{name: "invalid", input: "tokio", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCourse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}
