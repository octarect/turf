package model

import (
	"encoding/json"
	"time"
)

// RaceResult holds the full result of a finished race.
// It references the RaceCard and contains weather, going, post time,
// lap times, corner formations, and the list of Entry rows.
type RaceResult struct {
	RaceCard *RaceCard `json:"raceCard"`

	Going            Going             `json:"going"`
	Weather          Weather           `json:"weather"`
	PostTime         time.Time         `json:"postTime"`
	Entries          []Entry           `json:"entries"`
	LapTimes         []float64         `json:"lapTimes"`
	CornerFormations []CornerFormation `json:"cornerFormations"`
}

type Weather int

const (
	WeatherFine Weather = iota
	WeatherCloudy
	WeatherDrizzle
	WeatherRainy
)

func (w Weather) String() string {
	switch w {
	case WeatherFine:
		return "fine"
	case WeatherCloudy:
		return "cloudy"
	case WeatherDrizzle:
		return "drizzle"
	case WeatherRainy:
		return "rainy"
	}

	return "invalid"
}

func (w Weather) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

// Going represents the track condition (going) for a race.
// Turf and dirt values are interleaved using the formula iota/2*10 + iota%2,
// so that even iota values map to turf (0, 10, 20, 30) and odd values map to dirt (1, 11, 21, 31).
// This layout allows converting a turf going to its dirt equivalent by simply adding 1.
type Going int

const (
	GoingTurfGoodToFirm Going = iota/2*10 + iota%2
	GoingDirtStandard
	GoingTurfGood
	GoingDirtGood
	GoingTurfYielding
	GoingDirtMuddy
	GoingTurfSoft
	GoingDirtSloppy
)

func (g Going) String() string {
	switch g {
	case GoingTurfGoodToFirm:
		return "good_to_firm"
	case GoingDirtStandard:
		return "standard"
	case GoingTurfGood, GoingDirtGood:
		return "good"
	case GoingTurfYielding:
		return "yielding"
	case GoingDirtMuddy:
		return "muddy"
	case GoingTurfSoft:
		return "soft"
	case GoingDirtSloppy:
		return "sloppy"
	}

	return "invalid"
}

func (g Going) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

type FinishStatus int

const (
	FinishStatusNormal       FinishStatus = iota
	FinishStatusWithdrawn                 // 除外
	FinishStatusNonStarter                // 取消
	FinishStatusPulledUp                  // 中止
	FinishStatusDisqualified              // 失格
)

func (s FinishStatus) String() string {
	switch s {
	case FinishStatusNormal:
		return "normal"
	case FinishStatusWithdrawn:
		return "withdrawn"
	case FinishStatusNonStarter:
		return "non_starter"
	case FinishStatusPulledUp:
		return "pulled_up"
	case FinishStatusDisqualified:
		return "disqualified"
	}
	return "invalid"
}

func (s FinishStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Finish struct {
	Position int          `json:"position"`
	Status   FinishStatus `json:"status"`
}

type Entry struct {
	Finish          Finish           `json:"finish"`
	Bracket         int              `json:"bracket"`
	Num             int              `json:"num"`
	Weight          float64          `json:"weight"`
	FinishTime      float64          `json:"finishTime"`
	Margin          Margin           `json:"margin"`
	Last3F          float64          `json:"last3F"`
	CornerPositions []CornerPosition `json:"cornerPositions"`
	WinFavorite     int              `json:"winFavorite"`

	Horse   EntryHorse   `json:"horse"`
	Jockey  EntryJockey  `json:"jockey"`
	Trainer EntryTrainer `json:"trainer"`
}

type Margin struct {
	Kind   MarginKind `json:"kind"`
	Length float64    `json:"length"`
}

type MarginKind int

const (
	MarginKindLength MarginKind = iota
	MarginKindDeadHeat
	MarginKindNose
	MarginKindHead
	MarginKindNeck
	MarginKindDistance
)

func (k MarginKind) String() string {
	switch k {
	case MarginKindLength:
		return "length"
	case MarginKindDeadHeat:
		return "dead_heat"
	case MarginKindNose:
		return "nose"
	case MarginKindHead:
		return "head"
	case MarginKindNeck:
		return "neck"
	case MarginKindDistance:
		return "distance"
	}

	return "invalid"
}

func (k MarginKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

type EntryHorse struct {
	ID         HorseID  `json:"id"`
	Name       string   `json:"name"`
	NameEN     string   `json:"nameEN,omitempty"`
	Sex        HorseSex `json:"sex"`
	Age        int      `json:"age"`
	Weight     int      `json:"weight"`
	WeightDiff int      `json:"weightDiff"`
	CNAME      string   `json:"cname"`
}

type EntryJockey struct {
	Name      string `json:"name"`
	NameEN    string `json:"nameEN,omitempty"`
	Allowance int    `json:"allowance"`
	CNAME     string `json:"cname"`
}

type EntryTrainer struct {
	Name   string `json:"name"`
	NameEN string `json:"nameEN,omitempty"`
	CNAME  string `json:"cname"`
}

type CornerPosition struct {
	Corner   int `json:"corner"`
	Position int `json:"position"`
}

type CornerFormation struct {
	Corner    int    `json:"corner"`
	Formation string `json:"formation"`
}

type RaceResultTranslation struct {
	RaceName string
	Entries  map[HorseID]RaceResultTranslationEntry
}

type RaceResultTranslationEntry struct {
	HorseName   string
	JockeyName  string
	TrainerName string
}

func (r *RaceResult) ApplyTranslation(t *RaceResultTranslation) {
	if t == nil {
		return
	}
	if r.RaceCard != nil {
		r.RaceCard.SpecialNameEN = t.RaceName
	}
	for i := range r.Entries {
		entry := &r.Entries[i]
		if te, ok := t.Entries[entry.Horse.ID]; ok {
			entry.Horse.NameEN = te.HorseName
			entry.Jockey.NameEN = te.JockeyName
			entry.Trainer.NameEN = te.TrainerName
		}
	}
}
