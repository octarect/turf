package turf

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

type mockRaceCardService struct {
	raceCards []*model.RaceCard
	err       error
}

func (s *mockRaceCardService) ListRaceCards(ctx context.Context, q ListRaceCardsQuery) ([]*model.RaceCard, error) {
	return s.raceCards, s.err
}

type mockRaceResultGetter struct {
	result *model.RaceResult
	err    error
}

func (g *mockRaceResultGetter) GetRaceResult(ctx context.Context, raceCard *model.RaceCard) (*model.RaceResult, error) {
	return g.result, g.err
}

func TestRaceResultService_GetRaceResult(t *testing.T) {
	tests := []struct {
		name        string
		query       GetRaceResultQuery
		raceCards   []*model.RaceCard
		result      *model.RaceResult
		raceCardErr error
		getterErr   error
		want        *model.RaceResult
		wantErr     bool
	}{
		{
			name: "returns race result for resolved race card",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			raceCards: []*model.RaceCard{
				{
					SpecialName: "Tokyo 11R",
					Num:         11,
				},
			},
			result: &model.RaceResult{
				RaceCard: &model.RaceCard{
					SpecialName: "Tokyo 11R",
					Num:         11,
				},
			},
			want: &model.RaceResult{
				RaceCard: &model.RaceCard{
					SpecialName: "Tokyo 11R",
					Num:         11,
				},
			},
		},
		{
			name: "invalid when date is zero",
			query: GetRaceResultQuery{
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			wantErr: true,
		},
		{
			name: "invalid when course is unknown",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				RaceNo: 11,
			},
			wantErr: true,
		},
		{
			name: "invalid when race number is zero",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			wantErr: true,
		},
		{
			name: "invalid when date is before 1986-01-01",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "1985-12-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			wantErr: true,
		},
		{
			name: "returns error when no race card found for date course and race",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			wantErr: true,
		},
		{
			name: "returns error when multiple race cards found for date course and race",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			raceCards: []*model.RaceCard{
				{SpecialName: "Tokyo 11R A", Num: 11},
				{SpecialName: "Tokyo 11R B", Num: 11},
			},
			wantErr: true,
		},
		{
			name: "returns race card service error",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			raceCardErr: errors.New("race card service failed"),
			wantErr:     true,
		},
		{
			name: "returns race result getter error",
			query: GetRaceResultQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
				RaceNo: 11,
			},
			raceCards: []*model.RaceCard{
				{SpecialName: "Tokyo 11R", Num: 11},
			},
			getterErr: errors.New("race result getter failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRaceResultService(
				&mockRaceCardService{raceCards: tt.raceCards, err: tt.raceCardErr},
				&mockRaceResultGetter{result: tt.result, err: tt.getterErr},
				nil,
			)

			got, err := svc.GetRaceResult(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("GetRaceResult() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
