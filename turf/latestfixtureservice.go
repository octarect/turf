package turf

import (
	"context"
	"fmt"

	"github.com/octarect/turf/model"
)

type LatestFixtureService interface {
	ListLatestFixtures(context.Context, ListLatestFixturesQuery) ([]*model.Fixture, error)
}

type ListLatestFixturesQuery struct {
	Course model.Course
}

type latestFixtureService struct {
	lister LatestFixtureLister
}

var _ LatestFixtureService = &latestFixtureService{}

func NewLatestFixtureService(l LatestFixtureLister) *latestFixtureService {
	return &latestFixtureService{lister: l}
}

func (svc *latestFixtureService) ListLatestFixtures(ctx context.Context, q ListLatestFixturesQuery) ([]*model.Fixture, error) {
	fixtures, err := svc.lister.ListLatestFixtures(ctx)
	if err != nil {
		return nil, fmt.Errorf("list latest fixtures: %w", err)
	}

	if q.Course == model.CourseUnknown {
		return fixtures, nil
	}

	filtered := make([]*model.Fixture, 0, len(fixtures))
	for _, f := range fixtures {
		if f.Course == q.Course {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}
