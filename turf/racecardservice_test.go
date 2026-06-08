package turf

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/octarect/turf/model"
)

type mockFixtureService struct {
	fixtures []*model.Fixture
	err      error
}

func (s *mockFixtureService) ListFixtures(ctx context.Context, q ListFixturesQuery) ([]*model.Fixture, error) {
	return s.fixtures, s.err
}

type mockRaceCardLister struct {
	raceCards []*model.RaceCard
	err       error
}

func (l *mockRaceCardLister) ListRaceCards(ctx context.Context, fixture *model.Fixture) ([]*model.RaceCard, error) {
	return l.raceCards, l.err
}

func TestRaceCardService_ListRaceCards(t *testing.T) {
	tests := []struct {
		name       string
		query      ListRaceCardsQuery
		fixtures   []*model.Fixture
		raceCards  []*model.RaceCard
		fixtureErr error
		listerErr  error
		want       []*model.RaceCard
		wantErr    bool
	}{
		{
			name: "returns race cards for resolved fixture",
			query: ListRaceCardsQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "fixture-1",
				},
			},
			raceCards: []*model.RaceCard{
				{
					SpecialName: "Tokyo 11R",
					Num:         11,
				},
			},
			want: []*model.RaceCard{
				{
					SpecialName: "Tokyo 11R",
					Num:         11,
				},
			},
		},
		{
			name: "invalid when date is zero",
			query: ListRaceCardsQuery{
				Course: model.CourseTokyo,
			},
			wantErr: true,
		},
		{
			name: "invalid when course is unknown",
			query: ListRaceCardsQuery{
				Date: mustParseDate(t, "2024-05-31"),
			},
			wantErr: true,
		},
		{
			name: "invalid when date is before 1986-01-01",
			query: ListRaceCardsQuery{
				Date:   mustParseDate(t, "1985-12-31"),
				Course: model.CourseTokyo,
			},
			wantErr: true,
		},
		{
			name: "returns error when no fixture found for date and course",
			query: ListRaceCardsQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			wantErr: true,
		},
		{
			name: "returns error when multiple fixtures found for date and course",
			query: ListRaceCardsQuery{
				Date:   time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC),
				Course: model.CourseTokyo,
			},
			fixtures: []*model.Fixture{
				{Course: model.CourseTokyo, Date: model.Date{Time: mustParseDate(t, "2024-05-31")}, CNAME: "fixture-1"},
				{Course: model.CourseTokyo, Date: model.Date{Time: mustParseDate(t, "2024-05-31")}, CNAME: "fixture-2"},
			},
			wantErr: true,
		},
		{
			name: "returns fixture service error",
			query: ListRaceCardsQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			fixtureErr: errors.New("fixture service failed"),
			wantErr:    true,
		},
		{
			name: "returns race card lister error",
			query: ListRaceCardsQuery{
				Date:   mustParseDate(t, "2024-05-31"),
				Course: model.CourseTokyo,
			},
			fixtures: []*model.Fixture{
				{
					Course: model.CourseTokyo,
					Date:   model.Date{Time: mustParseDate(t, "2024-05-31")},
					CNAME:  "fixture-1",
				},
			},
			listerErr: errors.New("race card lister failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRaceCardService(
				&mockFixtureService{fixtures: tt.fixtures, err: tt.fixtureErr},
				&mockRaceCardLister{raceCards: tt.raceCards, err: tt.listerErr},
				nil,
			)

			got, err := svc.ListRaceCards(context.Background(), tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr = %v, error = %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("ListRaceCards() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
