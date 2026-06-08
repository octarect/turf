package turf

import (
	"context"
	"fmt"
	"time"

	"github.com/octarect/turf/model"
)

// RaceCardService is the service-level interface for listing race cards.
// It resolves the fixture internally from the query and applies English translation
// to specially-named races (e.g. G1 races) when a translator is available.
type RaceCardService interface {
	ListRaceCards(context.Context, ListRaceCardsQuery) ([]*model.RaceCard, error)
}

type ListRaceCardsQuery struct {
	Date   time.Time
	Course model.Course
}

func (q *ListRaceCardsQuery) Validate() error {
	if q.Date.IsZero() {
		return fmt.Errorf("date must be specified")
	}
	if q.Date.Before(earliestFixtureDate) {
		return fmt.Errorf("date must be on or after %s", earliestFixtureDate.Format("2006-01-02"))
	}
	if q.Course == model.CourseUnknown {
		return fmt.Errorf("course must be specified")
	}
	return nil
}

type raceCardService struct {
	fixtureSvc FixtureService
	lister     RaceCardLister
	translator RaceResultTranslator
}

var _ RaceCardService = &raceCardService{}

func NewRaceCardService(fixtureSvc FixtureService, lister RaceCardLister, translator RaceResultTranslator) *raceCardService {
	return &raceCardService{fixtureSvc: fixtureSvc, lister: lister, translator: translator}
}

func (svc *raceCardService) ListRaceCards(ctx context.Context, q ListRaceCardsQuery) ([]*model.RaceCard, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %v", err)
	}

	fixtures, err := svc.fixtureSvc.ListFixtures(ctx, ListFixturesQuery{
		Date:   q.Date,
		Course: q.Course,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve fixture: %w", err)
	}

	if len(fixtures) == 0 {
		return nil, fmt.Errorf("no fixture found for date=%s course=%v", q.Date.Format("2006-01-02"), q.Course)
	}
	if len(fixtures) > 1 {
		return nil, fmt.Errorf("multiple fixtures found for date=%s course=%v", q.Date.Format("2006-01-02"), q.Course)
	}

	raceCards, err := svc.lister.ListRaceCards(ctx, fixtures[0])
	if err != nil {
		return nil, fmt.Errorf("list race cards: %w", err)
	}

	if svc.translator != nil {
		for _, rc := range raceCards {
			if rc.SpecialName == "" {
				continue
			}
			translation, err := svc.translator.GetRaceResultTranslation(ctx, rc)
			if err == nil {
				rc.SpecialNameEN = translation.RaceName
			}
		}
	}

	return raceCards, nil
}
