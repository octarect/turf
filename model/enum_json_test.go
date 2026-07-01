package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEnumString(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "course", got: CourseTokyo.String(), want: "tokyo"},
		{name: "course unknown", got: CourseUnknown.String(), want: "unknown"},
		{name: "course invalid", got: Course(99).String(), want: "invalid"},
		{name: "age group 2yo", got: AgeGroup2.String(), want: "2yo"},
		{name: "age group 3yo", got: AgeGroup3.String(), want: "3yo"},
		{name: "age group 3yo+", got: AgeGroup3Plus.String(), want: "3yo_plus"},
		{name: "age group 4yo+", got: AgeGroup4Plus.String(), want: "4yo_plus"},
		{name: "age group 2yo JP", got: AgeGroup2.StringJP(), want: "2歳"},
		{name: "age group 3yo JP", got: AgeGroup3.StringJP(), want: "3歳"},
		{name: "age group 3yo+ JP", got: AgeGroup3Plus.StringJP(), want: "3歳以上"},
		{name: "age group 4yo+ JP", got: AgeGroup4Plus.StringJP(), want: "4歳以上"},
		{name: "grade newcomer", got: GradeNewComer.String(), want: "newcomer"},
		{name: "grade maiden", got: Grade0W.String(), want: "maiden"},
		{name: "grade invalid", got: Grade(99).String(), want: "invalid"},
		{name: "grade newcomer JP", got: GradeNewComer.StringJP(), want: "新馬"},
		{name: "grade maiden JP", got: Grade0W.StringJP(), want: "未勝利"},
		{name: "grade 1w JP", got: Grade1W.StringJP(), want: "1勝クラス"},
		{name: "surface unknown", got: SurfaceUnknown.String(), want: "unknown"},
		{name: "surface", got: SurfaceTurf.String(), want: "turf"},
		{name: "weather", got: WeatherCloudy.String(), want: "cloudy"},
		{name: "going", got: GoingTurfGoodToFirm.String(), want: "good_to_firm"},
		{name: "horse sex", got: HorseSexGelding.String(), want: "gelding"},
		{name: "margin kind", got: MarginKindDeadHeat.String(), want: "dead_heat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got = %q, want = %q", tt.got, tt.want)
			}
		})
	}
}

func TestEnumMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{name: "course", v: CourseTokyo, want: `"tokyo"`},
		{name: "grade newcomer", v: GradeNewComer, want: `"newcomer"`},
		{name: "grade", v: Grade0W, want: `"maiden"`},
		{name: "surface", v: SurfaceDirt, want: `"dirt"`},
		{name: "weather", v: WeatherRainy, want: `"rainy"`},
		{name: "going", v: GoingDirtMuddy, want: `"muddy"`},
		{name: "horse sex", v: HorseSexFemale, want: `"female"`},
		{name: "margin kind", v: MarginKindHead, want: `"head"`},
		{name: "invalid", v: Surface(99), want: `"invalid"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.v)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got = %s, want = %s", got, tt.want)
			}
		})
	}
}

func TestStructuredJSONUsesLabels(t *testing.T) {
	v := RaceResult{
		RaceCard: &RaceCard{
			SpecialName: "Tokyo Yushun",
			Num:         11,
			Grade:       GradeG1,
			AgeGroup:    AgeGroup3,
			Surface:     SurfaceTurf,
			Fixture: &Fixture{
				Course: CourseTokyo,
				Date:   Date{Time: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)},
			},
		},
		Going:   GoingTurfGoodToFirm,
		Weather: WeatherFine,
		Entries: []Entry{{
			Margin: Margin{Kind: MarginKindDeadHeat},
			Horse:  EntryHorse{Sex: HorseSexMale},
		}},
	}

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	got := string(b)
	for _, want := range []string{
		`"course":"tokyo"`,
		`"grade":"g1"`,
		`"ageGroup":"3yo"`,
		`"name":"Tokyo Yushun"`,
		`"surface":"turf"`,
		`"going":"good_to_firm"`,
		`"weather":"fine"`,
		`"kind":"dead_heat"`,
		`"sex":"male"`,
		`"raceCard":`,
		`"last3F":`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("JSON %s does not contain %s", got, want)
		}
	}
}

func TestRaceCardDisplayName(t *testing.T) {
	tests := []struct {
		name   string
		rc     RaceCard
		wantEN string
		wantJP string
	}{
		{
			name:   "special name",
			rc:     RaceCard{SpecialName: "日本ダービー", SpecialNameEN: "Japan Derby", AgeGroup: AgeGroup3, Grade: GradeG1},
			wantEN: "Japan Derby",
			wantJP: "日本ダービー",
		},
		{
			name:   "no special name falls back to generated",
			rc:     RaceCard{AgeGroup: AgeGroup3Plus, Grade: Grade1W},
			wantEN: "3yo_plus 1_win",
			wantJP: "3歳以上1勝クラス",
		},
		{
			name:   "special name EN missing falls back to generated",
			rc:     RaceCard{SpecialName: "日本ダービー", AgeGroup: AgeGroup3, Grade: GradeG1},
			wantEN: "3yo g1",
			wantJP: "日本ダービー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rc.DisplayName(); got != tt.wantEN {
				t.Fatalf("DisplayName() = %q, want %q", got, tt.wantEN)
			}
			if got := tt.rc.DisplayNameJP(); got != tt.wantJP {
				t.Fatalf("DisplayNameJP() = %q, want %q", got, tt.wantJP)
			}
		})
	}
}
