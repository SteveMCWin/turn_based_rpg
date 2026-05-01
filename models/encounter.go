package models

// used for populating rooms
type EncounterKind string

const (
	EncounterKindMonster EncounterKind = "monster"
	EncounterKindBoss    EncounterKind = "boss"
	EncounterKindEvent   EncounterKind = "event"
)

// If the encoutner if monster kind, event will be empty
type Encounter struct {
	Kind    EncounterKind `json:"kind"`
	Monster *Monster      `json:"monster,omitempty"`
	Event   *Event        `json:"event,omitempty"`
}

// Event is basically just something happened that affected your stats, could be good or bad
type Event struct {
	ID            string           `json:"id"`
	Description   string           `json:"description"`
	Applied       bool             `json:"applied,omitempty"`
	StatsAffected map[StatType]int `json:"stat_affected"`
}
