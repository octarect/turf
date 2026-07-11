package model

import "time"

func (r *RacePlan) ApplyTranslation(t *RaceResultTranslation) {
	if t == nil {
		return
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

type RacePlan struct {
	RaceCard *RaceCard `json:"raceCard"`

	FemaleOnly FemaleOnly      `json:"femaleOnly"`
	WeightRule WeightRule      `json:"weightRule"`
	PostTime   time.Time       `json:"postTime"`
	Entries    []RacePlanEntry `json:"entries"`
}

type RacePlanEntry struct {
	Bracket int     `json:"bracket"`
	Num     int     `json:"num"`
	Weight  float64 `json:"weight"`

	Horse   EntryHorse   `json:"horse"`
	Jockey  EntryJockey  `json:"jockey"`
	Trainer EntryTrainer `json:"trainer"`
}
