package turf

import (
	"context"
	"fmt"

	"github.com/octarect/turf/model"
)

type latestRaceCardService struct {
	latestFixtureSvc LatestFixtureService
	lister           RaceCardLister
}

var _ RaceCardService = &latestRaceCardService{}

func NewLatestRaceCardService(latestFixtureSvc LatestFixtureService, lister RaceCardLister) *latestRaceCardService {
	return &latestRaceCardService{latestFixtureSvc: latestFixtureSvc, lister: lister}
}

func (svc *latestRaceCardService) ListRaceCards(ctx context.Context, q ListRaceCardsQuery) ([]*model.RaceCard, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %v", err)
	}

	fixtures, err := svc.latestFixtureSvc.ListLatestFixtures(ctx, ListLatestFixturesQuery{Course: q.Course})
	if err != nil {
		return nil, fmt.Errorf("resolve fixture: %w", err)
	}

	var fixture *model.Fixture
	for _, f := range fixtures {
		if sameDate(f.Date.Time, q.Date) {
			fixture = f
			break
		}
	}
	if fixture == nil {
		return nil, fmt.Errorf("no fixture found for date=%s course=%v", q.Date.Format("2006-01-02"), q.Course)
	}

	raceCards, err := svc.lister.ListRaceCards(ctx, fixture)
	if err != nil {
		return nil, fmt.Errorf("list race cards: %w", err)
	}

	return raceCards, nil
}
