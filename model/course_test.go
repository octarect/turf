package model

import "testing"

func TestCourseByENName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Course
		wantErr bool
	}{
		{name: "lowercase", input: "tokyo", want: CourseTokyo},
		{name: "mixed case", input: "KyOtO", want: CourseKyoto},
		{name: "invalid", input: "foo", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CourseByENName(tt.input)
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
