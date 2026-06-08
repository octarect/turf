package jra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

func TestListCNAMESuffixMap(t *testing.T) {
	client, mux := setup(t)

	var html string
	mux.HandleFunc("/JRADB/accessS.html", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, html)
	})

	tests := []struct {
		name    string
		html    string
		want    CNAMESuffixMap
		wantErr bool
	}{
		{
			"success",
			`<script>var objParam=new Array();objParam["7001"]="AB";objParam["7002"]="CD";objParam["7003"]="EF";</script>`,
			CNAMESuffixMap{"7001": "AB", "7002": "CD", "7003": "EF"},
			false,
		},
		{
			"objParam not found.",
			"<script></script>",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html = tt.html
			got, err := client.listCNAMESuffixMap(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGenerateCNAME(t *testing.T) {
	client, mux := setup(t)

	var html string
	mux.HandleFunc("/JRADB/accessS.html", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, html)
	})

	now := time.Now().UTC()
	pastMonth := now.AddDate(0, -1, 0)
	if pastMonth.Format("0601") == now.Format("0601") {
		pastMonth = now.AddDate(0, -2, 0)
	}

	tests := []struct {
		name    string
		month   time.Time
		html    string
		want    string
		wantErr bool
	}{
		{
			name:  "past month",
			month: pastMonth,
			html:  `<script>var objParam=new Array();objParam["` + pastMonth.Format("0601") + `"]="AB";</script>`,
			want:  "pw01skl10" + pastMonth.Format("200601") + "/AB",
		},
		{
			name:  "current month",
			month: now,
			html:  `<script>var objParam=new Array();objParam["` + now.Format("0601") + `"]="CD";</script>`,
			want:  "pw01skl00" + now.Format("200601") + "/CD",
		},
		{
			name:    "missing suffix",
			month:   pastMonth,
			html:    `<script>var objParam=new Array();objParam["9999"]="ZZ";</script>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html = tt.html

			got, err := client.generateCNAME(context.Background(), tt.month)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("generateCNAME mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListFixtures(t *testing.T) {
	client, mux := setup(t)

	var suffixHTML string
	var fixtureHTML string
	mux.HandleFunc("/JRADB/accessS.html", func(w http.ResponseWriter, r *http.Request) {
		writeShiftJIS(t, w, suffixHTML)
	})
	mux.HandleFunc("/JRADB/accessR.html", func(w http.ResponseWriter, r *http.Request) {
		writeShiftJIS(t, w, fixtureHTML)
	})

	tests := []struct {
		name        string
		month       time.Time
		fixtureHTML string
		want        []*model.Fixture
		wantErr     bool
	}{
		{
			name:  "single",
			month: time.Date(2026, 1, 1, 0, 0, 0, 0, timeJST),
			fixtureHTML: fixturePageHTML(
				fixtureDateBlockHTML("1月31日",
					fixtureItemHTML("pw01abc123/AA", "1回東京1日"),
				),
			),
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Year:   2026,
					Season: 1,
					Day:    1,
					Date:   model.Date{Time: time.Date(2026, 1, 31, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01abc123/AA",
				},
			},
		},
		{
			name:  "multiple",
			month: time.Date(2024, 12, 1, 0, 0, 0, 0, timeJST),
			fixtureHTML: fixturePageHTML(
				fixtureDateBlockHTML("12月1日",
					fixtureItemHTML("pw01nakayama/AB", "5回中山2日"),
					fixtureItemHTML("pw01kyoto/CD", "7回京都2日"),
					fixtureItemHTML("pw01chukyo/EF", "4回中京2日"),
				),
				fixtureDateBlockHTML("12月28日",
					fixtureItemHTML("pw01nakayama/GH", "5回中山9日"),
					fixtureItemHTML("pw01kyoto/IJ", "7回京都9日"),
				),
			),
			want: []*model.Fixture{
				{
					Course: model.CourseNakayama,
					Year:   2024,
					Season: 5,
					Day:    2,
					Date:   model.Date{Time: time.Date(2024, 12, 1, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01nakayama/AB",
				},
				{
					Course: model.CourseKyoto,
					Year:   2024,
					Season: 7,
					Day:    2,
					Date:   model.Date{Time: time.Date(2024, 12, 1, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01kyoto/CD",
				},
				{
					Course: model.CourseChukyo,
					Year:   2024,
					Season: 4,
					Day:    2,
					Date:   model.Date{Time: time.Date(2024, 12, 1, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01chukyo/EF",
				},
				{
					Course: model.CourseNakayama,
					Year:   2024,
					Season: 5,
					Day:    9,
					Date:   model.Date{Time: time.Date(2024, 12, 28, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01nakayama/GH",
				},
				{
					Course: model.CourseKyoto,
					Year:   2024,
					Season: 7,
					Day:    9,
					Date:   model.Date{Time: time.Date(2024, 12, 28, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01kyoto/IJ",
				},
			},
		},
		{
			name:  "2-digit day",
			month: time.Date(2026, 5, 31, 0, 0, 0, 0, timeJST),
			fixtureHTML: fixturePageHTML(
				fixtureDateBlockHTML("5月31日",
					fixtureItemHTML("pw01abc123/AA", "2回東京12日"),
				),
			),
			want: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Year:   2026,
					Season: 2,
					Day:    12,
					Date:   model.Date{Time: time.Date(2026, 5, 31, 0, 0, 0, 0, timeJST)},
					CNAME:  "pw01abc123/AA",
				},
			},
		},
		{
			name:  "invalid fixture text",
			month: time.Date(2026, 5, 1, 0, 0, 0, 0, timeJST),
			fixtureHTML: fixturePageHTML(
				fixtureDateBlockHTML("5月3日",
					fixtureItemHTML("pw01abc123/AA", "東京1日"),
				),
			),
			wantErr: true,
		},
		{
			name:  "unknown course",
			month: time.Date(2026, 5, 1, 0, 0, 0, 0, timeJST),
			fixtureHTML: fixturePageHTML(
				fixtureDateBlockHTML("5月3日",
					fixtureItemHTML("pw01abc123/AA", "2回樺太1日"),
				),
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtureHTML = tt.fixtureHTML
			suffixHTML = fmt.Sprintf(`<script>var objParam=new Array();objParam[%q]="AA";</script>`, tt.month.Format("0601"))

			got, err := client.ListFixtures(context.Background(), tt.month)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("ListFixtures mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
