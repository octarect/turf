package jra

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

func TestListLatestFixtures(t *testing.T) {
	client, mux := setup(t)

	var html string
	mux.HandleFunc("/JRADB/accessD.html", func(w http.ResponseWriter, r *http.Request) {
		writeShiftJIS(t, w, html)
	})

	now := time.Now()
	year := now.Year()

	tests := []struct {
		name    string
		html    string
		want    []*model.Fixture
		wantErr bool
	}{
		{
			name: "single fixture",
			html: latestFixturesPageHTML(
				latestFixturesPanelHTML("7月12日",
					latestFixturesItemHTML("pw01abc123/AA", "2回東京1日"),
				),
			),
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Year:   year,
					Season: 2,
					Day:    1,
					Date:   model.Date{Time: time.Date(year, 7, 12, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01abc123/AA",
				},
			},
		},
		{
			name: "multiple fixtures across dates",
			html: latestFixturesPageHTML(
				latestFixturesPanelHTML("7月12日",
					latestFixturesItemHTML("pw01tokyo/AB", "2回東京1日"),
					latestFixturesItemHTML("pw01hanshin/CD", "3回阪神1日"),
				),
				latestFixturesPanelHTML("7月13日",
					latestFixturesItemHTML("pw01nakayama/EF", "3回中山2日"),
				),
			),
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Year:   year,
					Season: 2,
					Day:    1,
					Date:   model.Date{Time: time.Date(year, 7, 12, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01tokyo/AB",
				},
				{
					Course: model.CourseHanshin,
					Year:   year,
					Season: 3,
					Day:    1,
					Date:   model.Date{Time: time.Date(year, 7, 12, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01hanshin/CD",
				},
				{
					Course: model.CourseNakayama,
					Year:   year,
					Season: 3,
					Day:    2,
					Date:   model.Date{Time: time.Date(year, 7, 13, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01nakayama/EF",
				},
			},
		},
		{
			name:    "unknown course",
			html:    latestFixturesPageHTML(latestFixturesPanelHTML("7月12日", latestFixturesItemHTML("pw01abc/AA", "2回樺太1日"))),
			wantErr: true,
		},
		{
			name: "empty page",
			html: latestFixturesPageHTML(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html = tt.html

			got, err := client.ListLatestFixtures(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("ListLatestFixtures mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
