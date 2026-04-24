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
		floors[i].Rooms = make([]Room, rand.Intn(maxRoomsPerFloor)+1)
	}

	return floors
}

func FillFloorEncounters(floors []Floor, monsterTemplates []Monster, eventTemplates []Event) {
	monsters := slices.Clone(monsterTemplates)
	events := slices.Clone(eventTemplates)
	rand_monster_indexes := rand.Perm(len(monsters))
	rand_event_indexes := rand.Perm(len(events))

	for i := range floors {
		monster_room_idx := rand.Intn(len(floors[i].Rooms))
		max_room_connect_idx := 0

		for j := range floors[i].Rooms {
			floors[i].Rooms[j].ID = strconv.Itoa(i) + "," + strconv.Itoa(j)

			if j == monster_room_idx {
				floors[i].Rooms[j].Encounter.Kind = EncounterKindMonster
				floors[i].Rooms[j].Encounter.Monster = &monsters[rand_monster_indexes[len(rand_monster_indexes)-1]]
				rand_monster_indexes = rand_monster_indexes[:len(rand_monster_indexes)-1]
			} else {
				floors[i].Rooms[j].Encounter.Kind = EncounterKindEvent
				floors[i].Rooms[j].Encounter.Event = &events[rand_event_indexes[len(rand_event_indexes)-1]]
				rand_event_indexes = rand_event_indexes[:len(rand_event_indexes)-1]
			}

			if i < len(floors)-1 {
				nextFloorLen := len(floors[i+1].Rooms)
				isLastRoom := j == len(floors[i].Rooms)-1

				var room_indexes_to_connect int
				if isLastRoom {
					// Ensure all next-floor rooms up to the end are reachable
					room_indexes_to_connect = nextFloorLen - 1
				} else {
					room_indexes_to_connect = max(max_room_connect_idx, rand.Intn(nextFloorLen))
				}

				for room_idx := max_room_connect_idx; room_idx <= room_indexes_to_connect; room_idx++ {
					floors[i].Rooms[j].Next = append(floors[i].Rooms[j].Next, strconv.Itoa(i+1)+","+strconv.Itoa(room_idx))
				}
				max_room_connect_idx = room_indexes_to_connect + 1
			}
		}
	}
}
