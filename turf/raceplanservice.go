package turf

import (
	"context"
	"fmt"
	"time"

	"github.com/octarect/turf/model"
)

type RacePlanService interface {
	GetRacePlan(context.Context, GetRacePlanQuery) (*model.RacePlan, error)
}

type GetRacePlanQuery struct {
	Date   time.Time
	Course model.Course
	RaceNo int
}

func (q *GetRacePlanQuery) Validate() error {
	if q.Date.IsZero() {
		return fmt.Errorf("date must be specified")
	}
	if q.Course == model.CourseUnknown {
		return fmt.Errorf("course must be specified")
	}
	if q.RaceNo < 1 {
		return fmt.Errorf("race number must be greater than 0")
	}
	return nil
}

type racePlanService struct {
	latestFixtureSvc LatestFixtureService
	raceCardLister   RaceCardLister
	racePlanGetter   RacePlanGetter
	translator       RaceResultTranslator
}

var _ RacePlanService = &racePlanService{}

func NewRacePlanService(latestFixtureSvc LatestFixtureService, raceCardLister RaceCardLister, racePlanGetter RacePlanGetter, translator RaceResultTranslator) *racePlanService {
	return &racePlanService{
		latestFixtureSvc: latestFixtureSvc,
		raceCardLister:   raceCardLister,
		racePlanGetter:   racePlanGetter,
		translator:       translator,
	}
}

func (svc *racePlanService) GetRacePlan(ctx context.Context, q GetRacePlanQuery) (*model.RacePlan, error) {
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

	raceCards, err := svc.raceCardLister.ListRaceCards(ctx, fixture)
	if err != nil {
		return nil, fmt.Errorf("list race cards: %w", err)
	}

	var raceCard *model.RaceCard
	for _, rc := range raceCards {
		if rc.Num == q.RaceNo {
			raceCard = rc
			break
		}
	}
	if raceCard == nil {
		return nil, fmt.Errorf("race %d not found", q.RaceNo)
	}

	plan, err := svc.racePlanGetter.GetRacePlan(ctx, raceCard)
	if err != nil {
		return nil, fmt.Errorf("get race plan: %w", err)
	}

	if svc.translator != nil {
		t, err := svc.translator.GetRaceResultTranslation(ctx, raceCard)
		if err != nil {
			return nil, fmt.Errorf("get translation: %w", err)
		}
		plan.ApplyTranslation(t)
	}

	return plan, nil
}
