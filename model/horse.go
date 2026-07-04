package model

import "encoding/json"

type HorseID string
type JockeyID string
type TrainerID string

type HorseSex int

const (
	HorseSexMale HorseSex = iota
	HorseSexFemale
	HorseSexGelding
)

func (s HorseSex) String() string {
	switch s {
	case HorseSexMale:
		return "male"
	case HorseSexFemale:
		return "female"
	case HorseSexGelding:
		return "gelding"
	}

	return "invalid"
}

func (s HorseSex) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
