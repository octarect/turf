package jra

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

func TestRaceNameUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"normal", "東京優駿", "東京優駿", false},
		{"paren", "4歳以上1勝クラス(混合)", "4歳以上1勝クラス", false},
		{"bracket", "3歳未勝利[指定]", "3歳未勝利", false},
		{"paren and bracket", "2歳未勝利(混合)(指定)", "2歳未勝利", false},
		{"fullwidth paren", "障害4歳以上オープン（混合）", "障害4歳以上オープン", false},
		{"fullwidth bracket", "2歳新馬［指定］", "2歳新馬", false},
		{"fullwidth paren and bracket", "3歳以上オープン（国際）（特指）", "3歳以上オープン", false},
		{"mare(suffix)", "3歳未勝利牝［指定］", "3歳未勝利", false},
		{"mare(middle)", "府中牝馬S", "府中牝馬S", false},
		{"emperror", "天皇賞(春)", "天皇賞(春)", false},
		{"spaces", " 4歳以上3勝クラス[指定]", "4歳以上3勝クラス", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got raceName
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(raceName(tt.want), got); diff != "" {
				t.Fatalf("raceName mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDistanceUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    distance
		wantErr bool
	}{
		{"1600", "1600", distance(1600), false},
		{"2,400", "2,400", distance(2400), false},
		{"invalid", "abc", distance(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got distance
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

func TestSurfaceUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.Surface
		wantErr bool
	}{
		{"turf", "芝", model.SurfaceTurf, false},
		{"dirt", "ダート", model.SurfaceDirt, false},
		{"jump", "芝→ダート", model.SurfaceJump, false},
		{"invalid", "砂", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got surface
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(surface(tt.want), got); diff != "" {
				t.Fatalf("surface mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListRaceCards(t *testing.T) {
	client, mux := setup(t)

	var html string
	mux.HandleFunc("/JRADB/accessS.html", func(w http.ResponseWriter, r *http.Request) {
		writeShiftJIS(t, w, html)
	})

	fixture := &model.Fixture{
		Course: model.CourseTokyo,
		Year:   2026,
		Season: 2,
		Day:    1,
		Date:   model.Date{Time: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)},
		CNAME:  "pw01abc123",
	}

	tests := []struct {
		name    string
		html    string
		want    []*model.RaceCard
		wantErr bool
	}{
		{
			name: "single race",
			html: racePageHTML(
				raceRowHTML(1, "3歳未勝利", "", "", "1600", "芝", 16, "pw01race01"),
			),
			want: []*model.RaceCard{
				{
					SpecialName: "",
					Num:         1,
					Grade:       model.Grade0W,
					AgeGroup:    model.AgeGroup3,
					Surface:     model.SurfaceTurf,
					Distance:    1600,
					Runners:     16,
					CNAME:       "pw01race01",
					Fixture:     fixture,
				},
			},
		},
		{
			name: "newcomer race",
			html: racePageHTML(
				raceRowHTML(1, "メイクデビュー東京", "2歳新馬（混合）［指定］", "", "1200", "芝", 8, "pw01race01"),
			),
			want: []*model.RaceCard{
				{
					SpecialName: "",
					Num:         1,
					Grade:       model.GradeNewComer,
					AgeGroup:    model.AgeGroup2,
					Surface:     model.SurfaceTurf,
					Distance:    1200,
					Runners:     8,
					CNAME:       "pw01race01",
					Fixture:     fixture,
				},
			},
		},
		{
			name: "multiple races",
			html: racePageHTML(
				raceRowHTML(1, "3歳1勝クラス", "1勝クラス", "", "1800", "芝", 14, "pw01race01"),
				raceRowHTML(2, "4歳以上2勝クラス", "2勝クラス", "", "1200", "ダート", 12, "pw01race02"),
				raceRowHTML(11, "東京優駿", "3歳オープン（国際）牡・牝（指定）", "/img/icon_grade_s_g1.png", "2,400", "芝", 18, "pw01race11"),
			),
			want: []*model.RaceCard{
				{
					SpecialName: "3歳1勝クラス",
					Num:         1,
					Grade:       model.Grade1W,
					AgeGroup:    model.AgeGroup3,
					Surface:     model.SurfaceTurf,
					Distance:    1800,
					Runners:     14,
					CNAME:       "pw01race01",
					Fixture:     fixture,
				},
				{
					SpecialName: "4歳以上2勝クラス",
					Num:         2,
					Grade:       model.Grade2W,
					AgeGroup:    model.AgeGroup4Plus,
					Surface:     model.SurfaceDirt,
					Distance:    1200,
					Runners:     12,
					CNAME:       "pw01race02",
					Fixture:     fixture,
				},
				{
					SpecialName: "東京優駿",
					Num:         11,
					Grade:       model.GradeG1,
					AgeGroup:    model.AgeGroup3,
					Surface:     model.SurfaceTurf,
					Distance:    2400,
					Runners:     18,
					CNAME:       "pw01race11",
					Fixture:     fixture,
				},
			},
		},
		{
			name: "empty table",
			html: racePageHTML(),
			want: []*model.RaceCard{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html = tt.html

			got, err := client.ListRaceCards(context.Background(), fixture)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("ListRaceCards mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
