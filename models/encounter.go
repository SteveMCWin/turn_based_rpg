package models

type EncounterKind string

const (
	EncounterKindMonster EncounterKind = "monster"
	EncounterKindEvent   EncounterKind = "event"
)

type Encounter struct {
	Kind    EncounterKind `json:"kind"`
	Monster *Monster      `json:"monster,omitempty"`
	Event   *Event        `json:"event,omitempty"`
}

type Event struct {
	Description  string   `json:"description"`
	StatAffected StatType `json:"stat_affected"`
	Delta        int      `json:"delta"`
	Applied      bool     `json:"applied"`
}
