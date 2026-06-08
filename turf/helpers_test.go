package turf

import (
	"testing"
	"time"
)

func mustParseDate(t *testing.T, s string) time.Time {
	ret, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("failed to parse date: %v", err)
	}
	return ret
}
