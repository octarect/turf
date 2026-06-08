package model

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRaceResult_ApplyTranslation(t *testing.T) {
	tests := []struct {
		name        string
		result      *RaceResult
		translation *RaceResultTranslation
		want        *RaceResult
	}{
		{
			name: "applies translation to race card and entries",
			result: &RaceResult{
				RaceCard: &RaceCard{
					SpecialName: "日本ダービー",
				},
				Entries: []Entry{
					{
						Horse:   EntryHorse{ID: "2023100001", Name: "イクイノックス"},
						Jockey:  EntryJockey{Name: "川田将雅"},
						Trainer: EntryTrainer{Name: "木村哲也"},
					},
					{
						Horse:   EntryHorse{ID: "2023100002", Name: "ドウデュース"},
						Jockey:  EntryJockey{Name: "武豊"},
						Trainer: EntryTrainer{Name: "友道康夫"},
					},
				},
			},
			translation: &RaceResultTranslation{
				RaceName: "Japan Derby",
				Entries: map[HorseID]RaceResultTranslationEntry{
					"2023100001": {HorseName: "Equinox", JockeyName: "Y.Kawada", TrainerName: "T.Kimura"},
					"2023100002": {HorseName: "Do Deuce", JockeyName: "Y.Take", TrainerName: "Y.Tomodo"},
				},
			},
			want: &RaceResult{
				RaceCard: &RaceCard{
					SpecialName:   "日本ダービー",
					SpecialNameEN: "Japan Derby",
				},
				Entries: []Entry{
					{
						Horse:   EntryHorse{ID: "2023100001", Name: "イクイノックス", NameEN: "Equinox"},
						Jockey:  EntryJockey{Name: "川田将雅", NameEN: "Y.Kawada"},
						Trainer: EntryTrainer{Name: "木村哲也", NameEN: "T.Kimura"},
					},
					{
						Horse:   EntryHorse{ID: "2023100002", Name: "ドウデュース", NameEN: "Do Deuce"},
						Jockey:  EntryJockey{Name: "武豊", NameEN: "Y.Take"},
						Trainer: EntryTrainer{Name: "友道康夫", NameEN: "Y.Tomodo"},
					},
				},
			},
		},
		{
			name: "nil translation is no-op",
			result: &RaceResult{
				RaceCard: &RaceCard{SpecialName: "日本ダービー"},
				Entries: []Entry{
					{Horse: EntryHorse{ID: "2023100001", Name: "イクイノックス"}},
				},
			},
			translation: nil,
			want: &RaceResult{
				RaceCard: &RaceCard{SpecialName: "日本ダービー"},
				Entries: []Entry{
					{Horse: EntryHorse{ID: "2023100001", Name: "イクイノックス"}},
				},
			},
		},
		{
			name: "skips entries not present in translation",
			result: &RaceResult{
				RaceCard: &RaceCard{SpecialName: "日本ダービー"},
				Entries: []Entry{
					{Horse: EntryHorse{ID: "2023100001", Name: "イクイノックス"}},
					{Horse: EntryHorse{ID: "2023100002", Name: "ドウデュース"}},
				},
			},
			translation: &RaceResultTranslation{
				RaceName: "Japan Derby",
				Entries: map[HorseID]RaceResultTranslationEntry{
					"2023100001": {HorseName: "Equinox"},
				},
			},
			want: &RaceResult{
				RaceCard: &RaceCard{SpecialName: "日本ダービー", SpecialNameEN: "Japan Derby"},
				Entries: []Entry{
					{Horse: EntryHorse{ID: "2023100001", Name: "イクイノックス", NameEN: "Equinox"}},
					{Horse: EntryHorse{ID: "2023100002", Name: "ドウデュース"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.result.ApplyTranslation(tt.translation)
			if diff := cmp.Diff(tt.want, tt.result); diff != "" {
				t.Fatalf("ApplyTranslation() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
