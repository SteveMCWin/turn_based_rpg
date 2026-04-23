package models

import (
	"math/rand"
	"slices"
	"strconv"
)

// id of room is floor_num,room_num
type Room struct {
	ID          string
	Encounter   Encounter
	Next        []string
	IsCompleted bool
	CanEnter    bool
}

type Floor struct {
	Rooms []Room
	IsCompleted bool
}

func GenerateFloors(numFloors, maxRoomsPerFloor int) []Floor {
	floors := make([]Floor, numFloors)

	floors[0] = Floor{
		Rooms: []Room{Room{ CanEnter: true }},
		IsCompleted: false,
	}

	floors[numFloors-1] = Floor{
		Rooms: []Room{Room{}},
		IsCompleted: false,
	}

	for i := 1; i < numFloors-1; i++ {
		floors[i].Rooms = make([]Room, rand.Intn(3))
	}

	return floors
}

func FillFloorEncounters(floors []Floor, monsterTemplates []Monster, eventTemplates []Event) {
	monsters := slices.Clone(monsterTemplates)
	events := slices.Clone(eventTemplates)
	rand_monster_indexes := rand.Perm(len(monsters))
	rand_event_indexes := rand.Perm(len(events))

	for i, f := range floors {

		monster_room_idx := rand.Intn(len(f.Rooms))

		max_room_connect_idx := 0

		for j, r := range f.Rooms {
			r.ID = strconv.Itoa(i) + "," + strconv.Itoa(j)
			if j == monster_room_idx {
				r.Encounter.Kind = EncounterKindMonster
				r.Encounter.Monster = &monsters[rand_monster_indexes[len(rand_monster_indexes)-1]]
				rand_monster_indexes = rand_monster_indexes[:len(rand_monster_indexes)-1]
			} else {
				r.Encounter.Kind = EncounterKindEvent
				r.Encounter.Event = &events[rand_event_indexes[len(rand_event_indexes)-1]]
				rand_event_indexes = rand_event_indexes[:len(rand_event_indexes)-1]
			}

			if i < len(floors)-1 {
				room_indexes_to_connect := max(max_room_connect_idx, rand.Intn(len(floors[i+1].Rooms)))
				for room_idx := max_room_connect_idx; room_idx <= room_indexes_to_connect; room_idx++ {
					r.Next = append(r.Next, strconv.Itoa(i+1) + "," + strconv.Itoa(room_idx))
				}
				max_room_connect_idx = room_indexes_to_connect
			}

		}

	}
}
