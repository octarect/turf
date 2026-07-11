package jra

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

func TestGetRacePlan(t *testing.T) {
	client, mux := setup(t)

	var html string
	mux.HandleFunc("/JRADB/accessD.html", func(w http.ResponseWriter, r *http.Request) {
		writeShiftJIS(t, w, html)
	})

	fixture := &model.Fixture{
		Course: model.CourseTokyo,
		Year:   2026,
		Season: 2,
		Day:    1,
		Date:   model.Date{Time: time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)},
		CNAME:  "pw01dli00/XX",
	}

	raceCard := &model.RaceCard{
		Num:     11,
		CNAME:   "pw01dli11/YY",
		Fixture: fixture,
	}

	tests := []struct {
		name    string
		html    string
		want    *model.RacePlan
		wantErr bool
	}{
		{
			name: "single entry",
			html: racePlanPageHTML("10時15分",
				racePlanEntryHTML{
					Bracket:      1,
					HorseNo:      1,
					HorseName:    "テストホース",
					HorseCNAME:   "pw01hor000123456789/AA",
					SexAge:       "牡3/55.0",
					HorseWeight:  "500(+2)",
					Weight:       "55.0",
					JockeyName:   "武豊",
					JockeyCNAME:  "pw01jky000012345/AA",
					TrainerName:  "高橋文雅",
					TrainerCNAME: "pw01tri000067890/BB",
				},
			),
			want: &model.RacePlan{
				RaceCard:   raceCard,
				FemaleOnly: false,
				WeightRule: model.WeightRuleAge,
				PostTime:   time.Date(2026, 7, 12, 10, 15, 0, 0, timeJST),
				Entries: []model.RacePlanEntry{
					{
						Bracket: 1,
						Num:     1,
						Weight:  55.0,
						Horse: model.EntryHorse{
							ID:         "0123456789",
							Name:       "テストホース",
							Sex:        model.HorseSexMale,
							Age:        3,
							Weight:     500,
							WeightDiff: 2,
							CNAME:      "pw01hor000123456789/AA",
						},
						Jockey: model.EntryJockey{
							ID:        "12345",
							Name:      "武豊",
							Allowance: 0,
							CNAME:     "pw01jky000012345/AA",
						},
						Trainer: model.EntryTrainer{
							ID:    "67890",
							Name:  "高橋文雅",
							CNAME: "pw01tri000067890/BB",
						},
					},
				},
			},
		},
		{
			name: "empty entries",
			html: racePlanPageHTML("10時00分"),
			want: &model.RacePlan{
				RaceCard:   raceCard,
				FemaleOnly: false,
				WeightRule: model.WeightRuleAge,
				PostTime:   time.Date(2026, 7, 12, 10, 0, 0, 0, timeJST),
				Entries:    []model.RacePlanEntry{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html = tt.html

			got, err := client.GetRacePlan(context.Background(), raceCard)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("GetRacePlan mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
