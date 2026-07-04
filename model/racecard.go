package model

import "encoding/json"

// RaceCard is the entry list for a single race within a fixture.
// A fixture typically contains 12 race cards.
type RaceCard struct {
	SpecialName   string   `json:"-"`
	SpecialNameEN string   `json:"-"`
	Num           int      `json:"num"`
	Grade         Grade    `json:"grade"`
	AgeGroup      AgeGroup `json:"ageGroup"`
	Surface       Surface  `json:"surface"`
	Distance      int      `json:"distance"`
	Runners       int      `json:"runners"`
	// CNAME is an opaque identifier assigned by JRA to reference this race card.
	// It is extracted from href attributes in the HTML and used as a GET query parameter
	// to fetch the race result.
	CNAME   string   `json:"cname"`
	Fixture *Fixture `json:"fixture"`
}

func (rc *RaceCard) DisplayName() string {
	if rc.SpecialNameEN != "" {
		return rc.SpecialNameEN
	}
	s := rc.AgeGroup.String() + " " + rc.Grade.String()
	if rc.Surface == SurfaceJump {
		s += " jump"
	}
	return s
}

func (rc *RaceCard) DisplayNameJP() string {
	if rc.SpecialName != "" {
		return rc.SpecialName
	}
	s := rc.AgeGroup.StringJP() + rc.Grade.StringJP()
	if rc.Surface == SurfaceJump {
		s = "障害" + s
	}
	return s
}

func (rc *RaceCard) MarshalJSON() ([]byte, error) {
	type Alias RaceCard
	return json.Marshal(&struct {
		Name   string `json:"name"`
		NameEN string `json:"nameEN"`
		*Alias
	}{
		Name:   rc.DisplayNameJP(),
		NameEN: rc.DisplayName(),
		Alias:  (*Alias)(rc),
	})
}

type AgeGroup int

const (
	AgeGroup2 AgeGroup = iota
	AgeGroup3
	AgeGroup3Plus
	AgeGroup4Plus
)

func (a AgeGroup) String() string {
	switch a {
	case AgeGroup2:
		return "2yo"
	case AgeGroup3:
		return "3yo"
	case AgeGroup3Plus:
		return "3yo_plus"
	case AgeGroup4Plus:
		return "4yo_plus"
	}
	return "invalid"
}

func (a AgeGroup) StringJP() string {
	switch a {
	case AgeGroup2:
		return "2歳"
	case AgeGroup3:
		return "3歳"
	case AgeGroup3Plus:
		return "3歳以上"
	case AgeGroup4Plus:
		return "4歳以上"
	}
	return "Invalid"
}

func (a AgeGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

type Grade int

const (
	GradeNewComer Grade = iota
	Grade0W
	Grade1W
	Grade2W
	Grade3W
	GradeOP
	GradeG3
	GradeG2
	GradeG1
	GradeJG3
	GradeJG2
	GradeJG1
)

func (g Grade) String() string {
	switch g {
	case GradeNewComer:
		return "newcomer"
	case Grade0W:
		return "maiden"
	case Grade1W:
		return "1win"
	case Grade2W:
		return "2win"
	case Grade3W:
		return "3win"
	case GradeOP:
		return "open"
	case GradeG3:
		return "g3"
	case GradeG2:
		return "g2"
	case GradeG1:
		return "g1"
	case GradeJG3:
		return "jg3"
	case GradeJG2:
		return "jg2"
	case GradeJG1:
		return "jg1"
	}

	return "invalid"
}

func (g Grade) StringJP() string {
	switch g {
	case GradeNewComer:
		return "新馬"
	case Grade0W:
		return "未勝利"
	case Grade1W:
		return "1勝クラス"
	case Grade2W:
		return "2勝クラス"
	case Grade3W:
		return "3勝クラス"
	default:
		return "オープン"
	}
}

func (g Grade) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

type Surface int

const (
	SurfaceUnknown Surface = iota
	SurfaceTurf
	SurfaceDirt
	SurfaceJump
)

func (s Surface) String() string {
	switch s {
	case SurfaceUnknown:
		return "unknown"
	case SurfaceTurf:
		return "turf"
	case SurfaceDirt:
		return "dirt"
	case SurfaceJump:
		return "jump"
	}

	return "invalid"
}

func (s Surface) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
