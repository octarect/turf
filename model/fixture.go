package model

import (
	"time"
)

// Fixture represents a single racing day at a specific course.
// A fixture is identified by its course, year, season (e.g. "2回"), and day within the season (e.g. "1日").
// For example, "2回東京1日" means the first day of the second season at Tokyo racecourse.
type Fixture struct {
	Course Course `json:"course"`
	Year   int    `json:"year"`
	Season int    `json:"season"`
	Day    int    `json:"day"`
	Date   Date   `json:"date"`
	// CNAME is an opaque identifier assigned by JRA to reference this fixture.
	// It is extracted from onclick attributes in the HTML and used as a POST parameter
	// to fetch race cards for this fixture.
	CNAME string `json:"cname"`
}

func NewFixture(date time.Time, year, season, day int, courseName, cname string) (*Fixture, error) {
	course, err := getCourseByName(courseName)
	if err != nil {
		return nil, err
	}

	return &Fixture{
		Course: course,
		Year:   year,
		Season: season,
		Day:    day,
		Date:   Date{date},
		CNAME:  cname,
	}, nil
}

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}
