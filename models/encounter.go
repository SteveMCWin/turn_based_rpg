package models

type EncounterKind string

const (
	EncounterKindMonster EncounterKind = "monster"
	EncounterKindBoss    EncounterKind = "boss"
	EncounterKindEvent   EncounterKind = "event"
)

type Encounter struct {
	Kind    EncounterKind `json:"kind"`
	Monster *Monster      `json:"monster,omitempty"`
	Event   *Event        `json:"event,omitempty"`
}

type Event struct {
	ID            string           `json:"id"`
	Description   string           `json:"description"`
	Applied       bool             `json:"applied,omitempty"`
	StatsAffected map[StatType]int `json:"stat_affected"`
}
