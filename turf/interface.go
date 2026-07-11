package turf

import (
	"context"
	"time"

	"github.com/octarect/turf/model"
)

// FixtureLister is the low-level interface for fetching fixtures from JRA.
// It accepts a time.Time and returns all fixtures in the month containing that date.
// Implemented by scrape/jra.JRAClient.
type FixtureLister interface {
	ListFixtures(context.Context, time.Time) ([]*model.Fixture, error)
}

// RaceCardLister is the low-level interface for fetching race cards from JRA.
// It accepts a Fixture and returns all race cards belonging to it.
// Implemented by scrape/jra.JRAClient.
type RaceCardLister interface {
	ListRaceCards(context.Context, *model.Fixture) ([]*model.RaceCard, error)
}

// RaceResultGetter is the low-level interface for fetching a race result from JRA.
// It accepts a RaceCard and returns the full result using the card's CNAME.
// Implemented by scrape/jra.JRAClient.
type RaceResultGetter interface {
	GetRaceResult(context.Context, *model.RaceCard) (*model.RaceResult, error)
}

// RaceResultTranslator is the low-level interface for fetching English translations of race results.
// It accepts a RaceCard and returns English names for the race, horses, jockeys, and trainers.
// Implemented by scrape/jraen.JRAENClient.
type RaceResultTranslator interface {
	GetRaceResultTranslation(context.Context, *model.RaceCard) (*model.RaceResultTranslation, error)
}

// LatestFixtureLister is the low-level interface for fetching this week's fixtures from JRA.
// Implemented by scrape/jra.JRAClient.
type LatestFixtureLister interface {
	ListLatestFixtures(context.Context) ([]*model.Fixture, error)
}

// RacePlanGetter is the low-level interface for fetching a race plan (今週の出馬表) from JRA.
// It accepts a RaceCard and returns the full entry list.
// Implemented by scrape/jra.JRAClient.
type RacePlanGetter interface {
	GetRacePlan(context.Context, *model.RaceCard) (*model.RacePlan, error)
}
