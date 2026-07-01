package jra

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

func TestWeatherUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.Weather
		wantErr bool
	}{
		{"fine", "晴", model.WeatherFine, false},
		{"cloudy", "曇", model.WeatherCloudy, false},
		{"drizzle", "小雨", model.WeatherDrizzle, false},
		{"rainy", "雨", model.WeatherRainy, false},
		{"invalid", "雪", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got weather
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(weather(tt.want), got); diff != "" {
				t.Fatalf("weather mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGoingUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.Going
		wantErr bool
	}{
		{"good to firm", "良", model.GoingTurfGoodToFirm, false},
		{"good", "稍重", model.GoingTurfGood, false},
		{"yielding", "重", model.GoingTurfYielding, false},
		{"soft", "不良", model.GoingTurfSoft, false},
		{"invalid", "極悪", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got going
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(going(tt.want), got); diff != "" {
				t.Fatalf("going mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGoingOfSurface(t *testing.T) {
	tests := []struct {
		name    string
		going   going
		surface model.Surface
		want    model.Going
	}{
		{"turf keeps good to firm", going(model.GoingTurfGoodToFirm), model.SurfaceTurf, model.GoingTurfGoodToFirm},
		{"dirt maps good to firm to standard", going(model.GoingTurfGoodToFirm), model.SurfaceDirt, model.GoingDirtStandard},
		{"dirt maps good", going(model.GoingTurfGood), model.SurfaceDirt, model.GoingDirtGood},
		{"dirt maps yielding", going(model.GoingTurfYielding), model.SurfaceDirt, model.GoingDirtMuddy},
		{"dirt maps soft", going(model.GoingTurfSoft), model.SurfaceDirt, model.GoingDirtSloppy},
		{"unknown keeps turf value", going(model.GoingTurfGood), model.SurfaceUnknown, model.GoingTurfGood},
		{"jump keeps turf value", going(model.GoingTurfGood), model.SurfaceJump, model.GoingTurfGood},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.going.OfSurface(tt.surface)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("going mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFemaleOnlyUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  model.FemaleOnly
	}{
		{"female only", "牝", true},
		{"mixed", "混合", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got femaleOnly
			err := got.UnmarshalXPath([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(femaleOnly(tt.want), got); diff != "" {
				t.Fatalf("femaleOnly mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWeightRuleUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.WeightRule
		wantErr bool
	}{
		{"age", "馬齢", model.WeightRuleAge, false},
		{"special (別定)", "別定", model.WeightRuleSpecial, false},
		{"special (定量)", "定量", model.WeightRuleSpecial, false},
		{"handicap", "ハンデ", model.WeightRuleHandicap, false},
		{"with whitespace", "  馬齢  ", model.WeightRuleAge, false},
		{"invalid", "不明", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got weightRule
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(weightRule(tt.want), got); diff != "" {
				t.Fatalf("weightRule mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLapTimesUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    lapTimes
		wantErr bool
	}{
		{"hyphen-joined", "12.5 - 11.8 - 12.0", lapTimes{12.5, 11.8, 12.0}, false},
		{"single", "12.3", lapTimes{12.3}, false},
		{"invalid", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got lapTimes
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("lapTimes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHorseSexUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.HorseSex
		wantErr bool
	}{
		{"male", "牡", model.HorseSexMale, false},
		{"femare", "牝", model.HorseSexFemale, false},
		{"gelding", "セ", model.HorseSexGelding, false},
		{"invalid", "騸", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got horseSex
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(horseSex(tt.want), got); diff != "" {
				t.Fatalf("horseSex mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFinishTimeUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    finishTime
		wantErr bool
	}{
		{"1:23.4", "1:23.4", finishTime(83.4), false},
		{"invalid", "abc", finishTime(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got finishTime
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("finishTime mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFraction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"integer", "1", 1.0, false},
		{"fraction", "1/2", 0.5, false},
		{"mixed", "1 1/2", 1.5, false},
		{"fullwidth", "１ １/２", 1.5, false},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFraction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("parseFraction(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMarginUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    model.Margin
		wantErr bool
	}{
		{"empty (1st)", "", model.Margin{}, false},
		{"dead heat", "同着", model.Margin{Kind: model.MarginKindDeadHeat}, false},
		{"nose", "ハナ", model.Margin{Kind: model.MarginKindNose}, false},
		{"head", "アタマ", model.Margin{Kind: model.MarginKindHead}, false},
		{"neck", "クビ", model.Margin{Kind: model.MarginKindNeck}, false},
		{"distance", "大差", model.Margin{Kind: model.MarginKindDistance}, false},
		{"1 1/2", "1 1/2", model.Margin{Kind: model.MarginKindLength, Length: 1.5}, false},
		{"invalid", "???", model.Margin{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got margin
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(margin(tt.want), got); diff != "" {
				t.Fatalf("margin mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHorseWeightUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    horseWeight
		wantErr bool
	}{
		{"positive diff", "480(+2)", horseWeight{Weight: 480, Diff: 2}, false},
		{"negative diff", "480(-4)", horseWeight{Weight: 480, Diff: -4}, false},
		{"zero diff", "480(0)", horseWeight{Weight: 480, Diff: 0}, false},
		{"first run", "480(初出走)", horseWeight{Weight: 480, Diff: 0, Other: "初出走"}, false},
		{"no diff", "480", horseWeight{Weight: 480, Diff: 0}, false},
		{"invalid", "abc", horseWeight{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got horseWeight
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("horseWeight mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFinishPositionUnmarshalXPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    finishPosition
		wantErr bool
	}{
		{"normal", "1", finishPosition{Position: 1, Status: model.FinishStatusNormal}, false},
		{"withdrawn", "除外", finishPosition{Status: model.FinishStatusWithdrawn}, false},
		{"non-starter", "取消", finishPosition{Status: model.FinishStatusNonStarter}, false},
		{"pulled up", "中止", finishPosition{Status: model.FinishStatusPulledUp}, false},
		{"disqualified", "失格", finishPosition{Status: model.FinishStatusDisqualified}, false},
		{"invalid", "abc", finishPosition{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got finishPosition
			err := got.UnmarshalXPath([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("finishPosition mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetRaceResult(t *testing.T) {
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
		CNAME:  "pw01fix01",
	}

	raceCard := &model.RaceCard{
		SpecialName: "東京優駿",
		Num:         11,
		Grade:       model.GradeG1,
		AgeGroup:    model.AgeGroup3,
		Surface:     model.SurfaceTurf,
		Distance:    2400,
		Runners:     18,
		CNAME:       "pw01race11",
		Fixture:     fixture,
	}

	html = raceResultPageHTML(resultOpts{
		Weather:    "晴",
		GoingClass: "turf",
		GoingText:  "良",
		FemaleOnly: "混合",
		WeightRule: "馬齢",
		PostTime:   "15時40分",
		LapTimes:   "12.5 - 11.8 - 12.0",
		CornerFormations: []cornerFormationHTML{
			{Corner: 3, Formation: "5-10-2"},
			{Corner: 4, Formation: "2-5-10"},
		},
		Entries: []entryHTML{
			{
				FinishPosition: "1",
				Bracket:        5,
				HorseNo:        10,
				HorseName:      "テストホース1",
				HorseCNAME:     "pw01dby012023000001/A1",
				SexAge:         "牡3",
				Weight:         "57",
				JockeyName:     "テスト騎手",
				JockeyCNAME:    "pw01jockey01",
				FinishTime:     "2:24.1",
				Margin:         "",
				Corners: []cornerPositionHTML{
					{Corner: 3, Position: 5},
					{Corner: 4, Position: 2},
				},
				Last3F:       "34.5",
				HorseWeight:  "480(+2)",
				TrainerName:  "テスト調教師",
				TrainerCNAME: "pw01trainer01",
				WinFavorite:  1,
			},
			{
				FinishPosition:  "2",
				Bracket:         3,
				HorseNo:         5,
				HorseName:       "テストホース2",
				HorseCNAME:      "pw01dby012023000002/B2",
				SexAge:          "牝4",
				Weight:          "55",
				JockeyName:      "減量テスト騎手",
				JockeyAllowance: 3,
				JockeyCNAME:     "pw01jockey02",
				FinishTime:      "2:24.5",
				Margin:          "クビ",
				Corners: []cornerPositionHTML{
					{Corner: 3, Position: 3},
					{Corner: 4, Position: 4},
				},
				Last3F:       "34.8",
				HorseWeight:  "460(0)",
				TrainerName:  "テスト調教師2",
				TrainerCNAME: "pw01trainer02",
				WinFavorite:  3,
			},
			{
				FinishPosition: "3",
				Bracket:        7,
				HorseNo:        14,
				HorseName:      "テストホース3",
				HorseCNAME:     "pw01dby012023000003/C3",
				SexAge:         "牡5",
				Weight:         "57",
				JockeyName:     "リンク無し騎手",
				JockeyCNAME:    "",
				FinishTime:     "2:25.0",
				Margin:         "１ 1/2",
				Corners: []cornerPositionHTML{
					{Corner: 3, Position: 8},
					{Corner: 4, Position: 7},
				},
				Last3F:       "35.1",
				HorseWeight:  "500(-4)",
				TrainerName:  "テスト調教師3",
				TrainerCNAME: "pw01trainer03",
				WinFavorite:  5,
			},
		},
	})

	got, err := client.GetRaceResult(context.Background(), raceCard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timeJST := time.FixedZone("Asia/Tokyo", 9*60*60)
	want := &model.RaceResult{
		RaceCard:   raceCard,
		Going:      model.GoingTurfGoodToFirm,
		Weather:    model.WeatherFine,
		FemaleOnly: false,
		WeightRule: model.WeightRuleAge,
		PostTime:   time.Date(2026, 5, 3, 15, 40, 0, 0, timeJST),
		LapTimes:   []float64{12.5, 11.8, 12.0},
		CornerFormations: []model.CornerFormation{
			{Corner: 3, Formation: "5-10-2"},
			{Corner: 4, Formation: "2-5-10"},
		},
		Entries: []model.Entry{
			{
				Finish:     model.Finish{Position: 1, Status: model.FinishStatusNormal},
				Bracket:    5,
				Num:        10,
				Weight:     57,
				FinishTime: 144.1,
				Margin:     model.Margin{},
				Last3F:     34.5,
				CornerPositions: []model.CornerPosition{
					{Corner: 3, Position: 5},
					{Corner: 4, Position: 2},
				},
				WinFavorite: 1,
				Horse: model.EntryHorse{
					Name:       "テストホース1",
					Sex:        model.HorseSexMale,
					Age:        3,
					Weight:     480,
					WeightDiff: 2,
					ID:         "2023000001",
					CNAME:      "pw01dby012023000001/A1",
				},
				Jockey: model.EntryJockey{
					Name:  "テスト騎手",
					CNAME: "pw01jockey01",
				},
				Trainer: model.EntryTrainer{
					Name:  "テスト調教師",
					CNAME: "pw01trainer01",
				},
			},
			{
				Finish:     model.Finish{Position: 2, Status: model.FinishStatusNormal},
				Bracket:    3,
				Num:        5,
				Weight:     55,
				FinishTime: 144.5,
				Margin:     model.Margin{Kind: model.MarginKindNeck},
				Last3F:     34.8,
				CornerPositions: []model.CornerPosition{
					{Corner: 3, Position: 3},
					{Corner: 4, Position: 4},
				},
				WinFavorite: 3,
				Horse: model.EntryHorse{
					Name:       "テストホース2",
					Sex:        model.HorseSexFemale,
					Age:        4,
					Weight:     460,
					WeightDiff: 0,
					ID:         "2023000002",
					CNAME:      "pw01dby012023000002/B2",
				},
				Jockey: model.EntryJockey{
					Name:      "減量テスト騎手",
					Allowance: 3,
					CNAME:     "pw01jockey02",
				},
				Trainer: model.EntryTrainer{
					Name:  "テスト調教師2",
					CNAME: "pw01trainer02",
				},
			},
			{
				Finish:     model.Finish{Position: 3, Status: model.FinishStatusNormal},
				Bracket:    7,
				Num:        14,
				Weight:     57,
				FinishTime: 145.0,
				Margin:     model.Margin{Kind: model.MarginKindLength, Length: 1.5},
				Last3F:     35.1,
				CornerPositions: []model.CornerPosition{
					{Corner: 3, Position: 8},
					{Corner: 4, Position: 7},
				},
				WinFavorite: 5,
				Horse: model.EntryHorse{
					Name:       "テストホース3",
					Sex:        model.HorseSexMale,
					Age:        5,
					Weight:     500,
					WeightDiff: -4,
					ID:         "2023000003",
					CNAME:      "pw01dby012023000003/C3",
				},
				Jockey: model.EntryJockey{
					Name:  "リンク無し騎手",
					CNAME: "",
				},
				Trainer: model.EntryTrainer{
					Name:  "テスト調教師3",
					CNAME: "pw01trainer03",
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("GetRaceResult mismatch (-want +got):\n%s", diff)
	}
}
