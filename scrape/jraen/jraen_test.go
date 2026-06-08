package jraen

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
	"github.com/octarect/turf/turf"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func setup(t *testing.T) (*JRAENClient, *http.ServeMux) {
	t.Helper()
	tc := turf.NewClient(&http.Client{Timeout: 30 * time.Second})
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL + "/")
	tc.SetBaseURL(u)
	return NewJRAENClient(tc), mux
}

func writeShiftJIS(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	writer := transform.NewWriter(w, japanese.ShiftJIS.NewEncoder())
	defer writer.Close()
	if _, err := io.WriteString(writer, body); err != nil {
		t.Fatalf("failed to write Shift-JIS response: %v", err)
	}
}

type jraenEntryHTML struct {
	HorseID     string
	HorseName   string
	JockeyName  string
	TrainerName string
}

func jraenResultPageHTML(raceName string, entries []jraenEntryHTML) string {
	var rows strings.Builder
	for _, e := range entries {
		rows.WriteString(`<tr>`)
		rows.WriteString(`<td class="raceHorse"><a horseno="` + e.HorseID + `" bamei="` + e.HorseName + `">link</a></td>`)
		rows.WriteString(`<td class="raceHorse"></td>`)
		rows.WriteString(`<td class="raceHorse"></td>`)
		rows.WriteString(`<td class="raceHorse">` + e.JockeyName + `<br/>` + e.TrainerName + `</td>`)
		rows.WriteString(`</tr>`)
	}
	return `<body>` +
		`<table><tr><td>header</td></tr><tr><td>` + raceName + `</td></tr></table>` +
		`<table class="running"></table>` +
		`<table class="running"><tr><td>header row</td></tr>` + rows.String() + `</table>` +
		`</body>`
}

func newRaceCard() *model.RaceCard {
	return &model.RaceCard{
		Num:      11,
		Grade:    model.GradeG1,
		AgeGroup: model.AgeGroup3,
		Surface:  model.SurfaceTurf,
		Fixture: &model.Fixture{
			Course: model.CourseTokyo,
			Year:   2025,
			Season: 3,
			Day:    7,
			Date:   model.Date{Time: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
}

func TestGetRaceResultTranslation(t *testing.T) {
	tests := []struct {
		name     string
		raceName string
		entries  []jraenEntryHTML
		want     *model.RaceResultTranslation
	}{
		{
			name:     "normal",
			raceName: "Japan Derby",
			entries: []jraenEntryHTML{
				{HorseID: "2023000001", HorseName: "Equus One", JockeyName: "Jockey A", TrainerName: "Trainer A"},
				{HorseID: "2023000002", HorseName: "Equus Two", JockeyName: "Jockey B", TrainerName: "Trainer B"},
			},
			want: &model.RaceResultTranslation{
				RaceName: "Japan Derby",
				Entries: map[model.HorseID]model.RaceResultTranslationEntry{
					"2023000001": {HorseName: "Equus One", JockeyName: "Jockey A", TrainerName: "Trainer A"},
					"2023000002": {HorseName: "Equus Two", JockeyName: "Jockey B", TrainerName: "Trainer B"},
				},
			},
		},
		{
			name:     "grade_suffix_stripped",
			raceName: "Japan Derby(G1)",
			entries:  []jraenEntryHTML{},
			want: &model.RaceResultTranslation{
				RaceName: "Japan Derby",
				Entries:  map[model.HorseID]model.RaceResultTranslationEntry{},
			},
		},
		{
			name:     "horse_country_stripped",
			raceName: "Some Race",
			entries: []jraenEntryHTML{
				{HorseID: "2023000001", HorseName: "Foreign Horse(USA)", JockeyName: "Jockey A", TrainerName: "Trainer A"},
			},
			want: &model.RaceResultTranslation{
				RaceName: "Some Race",
				Entries: map[model.HorseID]model.RaceResultTranslationEntry{
					"2023000001": {HorseName: "Foreign Horse", JockeyName: "Jockey A", TrainerName: "Trainer A"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mux := setup(t)
			mux.HandleFunc("/kaisai/running", func(w http.ResponseWriter, r *http.Request) {
				writeShiftJIS(t, w, jraenResultPageHTML(tt.raceName, tt.entries))
			})

			got, err := client.GetRaceResultTranslation(context.Background(), newRaceCard())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
