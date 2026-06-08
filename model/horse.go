package model

import "encoding/json"

type HorseID string

type HorseSex int

const (
	HorseSexMale HorseSex = iota
	HorseSexMare
	HorseSexGelding
)

func (s HorseSex) String() string {
	switch s {
	case HorseSexMale:
		return "male"
	case HorseSexMare:
		return "mare"
	case HorseSexGelding:
		return "gelding"
	}

	return "invalid"
}

func (s HorseSex) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
