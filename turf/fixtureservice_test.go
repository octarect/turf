package turf

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

type mockFIxtureLister struct {
	fixtures []*model.Fixture
}

func (l *mockFIxtureLister) ListFixtures(ctx context.Context, month time.Time) ([]*model.Fixture, error) {
	return l.fixtures, nil
}

func TestFixtureService_ListFixtures(t *testing.T) {
	tests := []struct {
		name     string
		query    ListFixturesQuery
		fixtures []*model.Fixture
		want     []*model.Fixture
		wantErr  bool
	}{
		{
			name: "monthly",
			query: ListFixturesQuery{
				Month: mustParseDate(t, "2024-06-01"),
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-01")},
					CNAME:  "tokyo-0601",
				},
				{
					Course: model.CourseKyoto,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-02")},
					CNAME:  "kyoto-0602",
				},
			},
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-01")},
					CNAME:  "tokyo-0601",
				},
				{
					Course: model.CourseKyoto,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-02")},
					CNAME:  "kyoto-0602",
				},
			},
			wantErr: false,
		},
		{
			name: "daily",
			query: ListFixturesQuery{
				Date: mustParseDate(t, "2024-05-31"),
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "tokyo-0531",
				},
				{
					Course: model.CourseKyoto,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-25")},
					CNAME:  "kyoto-0525",
				},
			},
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "tokyo-0531",
				},
			},
			wantErr: false,
		},
		{
			name: "monthly with course filter",
			query: ListFixturesQuery{
				Month:  mustParseDate(t, "2024-06-01"),
				Course: model.CourseTokyo,
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-01")},
					CNAME:  "tokyo-0601",
				},
				{
					Course: model.CourseKyoto,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-01")},
					CNAME:  "kyoto-0601",
				},
			},
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-06-01")},
					CNAME:  "tokyo-0601",
				},
			},
			wantErr: false,
		},
		{
			name: "daily with course filter",
			query: ListFixturesQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "tokyo-0531",
				},
				{
					Course: model.CourseKyoto,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "kyoto-0531",
				},
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-25")},
					CNAME:  "tokyo-0525",
				},
			},
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "tokyo-0531",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid when neither month nor date is specified",
			query:   ListFixturesQuery{},
			wantErr: true,
		},
		{
			name: "invalid when both month and date are specified",
			query: ListFixturesQuery{
				Month: mustParseDate(t, "2024-06-01"),
				Date:  mustParseDate(t, "2024-06-01"),
			},
			wantErr: true,
		},
		{
			name: "invalid when month is before 1986-01",
			query: ListFixturesQuery{
				Month: mustParseDate(t, "1985-12-01"),
			},
			wantErr: true,
		},
		{
			name: "invalid when date is before 1986-01-01",
			query: ListFixturesQuery{
				Date: mustParseDate(t, "1985-12-31"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewFixtureService(&mockFIxtureLister{fixtures: tt.fixtures})
			got, err := svc.ListFixtures(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("ListFixtures() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
