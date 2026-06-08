package turf

import (
	"context"
	"fmt"
	"time"

	"github.com/octarect/turf/model"
)

// FixtureService is the service-level interface for listing fixtures.
// Unlike FixtureLister, it accepts a query object and handles filtering by date and course.
type FixtureService interface {
	ListFixtures(context.Context, ListFixturesQuery) ([]*model.Fixture, error)
}

// ListFixturesQuery is the input to FixtureService.ListFixtures.
// Exactly one of Month or Date must be set; both cannot be set simultaneously.
// Course is optional; when set to CourseUnknown, results are not filtered by course.
type ListFixturesQuery struct {
	Month  time.Time
	Date   time.Time
	Course model.Course
}

var earliestFixtureDate = time.Date(1986, 1, 1, 0, 0, 0, 0, time.UTC)

// EarliestFixtureDate returns the earliest date for which JRA fixture data is available.
// JRA's online database starts from 1986-01-01.
func EarliestFixtureDate() time.Time {
	return earliestFixtureDate
}

func (q *ListFixturesQuery) Validate() error {
	if !q.Month.IsZero() && !q.Date.IsZero() {
		return fmt.Errorf("only one of Month or Date can be specified")
	}
	if q.Month.IsZero() && q.Date.IsZero() {
		return fmt.Errorf("either Month or Date must be specified")
	}
	if !q.Month.IsZero() && q.Month.Before(earliestFixtureDate) {
		return fmt.Errorf("month must be on or after %s", earliestFixtureDate.Format("2006-01"))
	}
	if !q.Date.IsZero() && q.Date.Before(earliestFixtureDate) {
		return fmt.Errorf("date must be on or after %s", earliestFixtureDate.Format("2006-01-02"))
	}
	return nil
}

type fixtureService struct {
	lister FixtureLister
}

var _ FixtureService = &fixtureService{}

func NewFixtureService(l FixtureLister) *fixtureService {
	return &fixtureService{lister: l}
}

func (svc *fixtureService) ListFixtures(ctx context.Context, q ListFixturesQuery) ([]*model.Fixture, error) {
	if err := q.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %v", err)
	}

	t := q.Month
	if t.IsZero() {
		t = q.Date
	}

	fixtures, err := svc.lister.ListFixtures(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("list fixtures: %w", err)
	}

	filtered := make([]*model.Fixture, 0, len(fixtures))
	for _, fixture := range fixtures {
		if !q.Date.IsZero() && !sameDate(fixture.Date.Time, q.Date) {
			continue
		}
		if q.Course != model.CourseUnknown && fixture.Course != q.Course {
			continue
		}
		filtered = append(filtered, fixture)
	}

	return filtered, nil
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
