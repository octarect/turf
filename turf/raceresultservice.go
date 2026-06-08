package turf

import (
	"context"
	"fmt"
	"time"

	"github.com/octarect/turf/model"
)

// RaceResultService is the service-level interface for fetching a race result.
// It resolves the fixture and race card internally from the query,
// and applies English translation when a translator is available.
type RaceResultService interface {
	GetRaceResult(context.Context, GetRaceResultQuery) (*model.RaceResult, error)
}

type GetRaceResultQuery struct {
	Date   time.Time
	Course model.Course
	RaceNo int
}

func (q *GetRaceResultQuery) Validate() error {
	if q.Date.IsZero() {
		return fmt.Errorf("date must be specified")
	}
	if q.Date.Before(earliestFixtureDate) {
		return fmt.Errorf("date must be on or after %s", earliestFixtureDate.Format("2006-01-02"))
	}
	if q.Course == model.CourseUnknown {
		return fmt.Errorf("course must be specified")
	}
	if q.RaceNo < 1 {
		return fmt.Errorf("race number must be greater than 0")
	}
	return nil
}

type raceResultService struct {
	raceCardSvc RaceCardService
	getter      RaceResultGetter
	translator  RaceResultTranslator
}

var _ RaceResultService = &raceResultService{}

func NewRaceResultService(raceCardSvc RaceCardService, getter RaceResultGetter, translator RaceResultTranslator) *raceResultService {
	return &raceResultService{raceCardSvc: raceCardSvc, getter: getter, translator: translator}
}

func (svc *raceResultService) GetRaceResult(ctx context.Context, q GetRaceResultQuery) (*model.RaceResult, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %v", err)
	}

	raceCards, err := svc.raceCardSvc.ListRaceCards(ctx, ListRaceCardsQuery{
		Date:   q.Date,
		Course: q.Course,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve race cards: %w", err)
	}

	var matched []*model.RaceCard
	for _, raceCard := range raceCards {
		if raceCard.Num != q.RaceNo {
			continue
		}
		matched = append(matched, raceCard)
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no race card found for date=%s course=%v race=%d", q.Date.Format("2006-01-02"), q.Course, q.RaceNo)
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("multiple race cards found for date=%s course=%v race=%d", q.Date.Format("2006-01-02"), q.Course, q.RaceNo)
	}

	result, err := svc.getter.GetRaceResult(ctx, matched[0])
	if err != nil {
		return nil, fmt.Errorf("get race result: %w", err)
	}

	if svc.translator != nil {
		translation, err := svc.translator.GetRaceResultTranslation(ctx, matched[0])
		if err == nil {
			result.ApplyTranslation(translation)
		}
	}

	return result, nil
}
