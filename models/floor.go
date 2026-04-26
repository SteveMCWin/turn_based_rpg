package models

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
)

const ROOM_ID_SEPARATOR string = ","

// id of room is floor_num,room_num
type Room struct {
	ID          string
	Encounter   Encounter
	NextRoomIDs []string
	IsCompleted bool
	CanEnter    bool
}

type Floor struct {
	Rooms       []Room
	IsCompleted bool
}

func GenerateFloors(numFloors, maxRoomsPerFloor int) []Floor {
	floors := make([]Floor, numFloors)

	floors[0] = Floor{
		Rooms:       []Room{Room{CanEnter: true}},
		IsCompleted: false,
	}

	floors[numFloors-1] = Floor{
		Rooms:       []Room{Room{}},
		IsCompleted: false,
	}

	for i := 1; i < numFloors-1; i++ {
		floors[i].Rooms = make([]Room, rand.Intn(maxRoomsPerFloor)+1)
	}

	return floors
}

func GetFloorIdx(room_id string) (int, int, error) {
	idxs := strings.Split(room_id, ROOM_ID_SEPARATOR)
	if len(idxs) != 2 {
		return -1, -1, fmt.Errorf("Error getting floor and room idx from room id %s\n", room_id)
	}

	f_idx, err := strconv.Atoi(idxs[0])
	if err != nil {
		return -1, -1, fmt.Errorf("Error converting string to floor idx: %s", err)
	}

	r_idx, err := strconv.Atoi(idxs[1])
	if err != nil {
		return -1, -1, fmt.Errorf("Error converting string to room idx: %s", err)
	}

	return f_idx, r_idx, nil
}

func FillFloorEncounters(floors []Floor, monsters []Monster, eventTemplates []Event) {
	events := slices.Clone(eventTemplates)
	rand_monster_indexes := rand.Perm(len(monsters))
	rand_event_indexes := rand.Perm(len(events))

	for i := range floors {
		monster_room_idx := rand.Intn(len(floors[i].Rooms))
		max_room_connect_idx := 0

		for j := range floors[i].Rooms {
			floors[i].Rooms[j].ID = strconv.Itoa(i) + ROOM_ID_SEPARATOR + strconv.Itoa(j)

			if j == monster_room_idx {
				floors[i].Rooms[j].Encounter.Kind = EncounterKindMonster
				floors[i].Rooms[j].Encounter.Monster = &monsters[rand_monster_indexes[len(rand_monster_indexes)-1]]
				rand_monster_indexes = rand_monster_indexes[:len(rand_monster_indexes)-1]

				floors[i].Rooms[j].Encounter.Monster.SetToLevel(i+1)

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
					floors[i].Rooms[j].NextRoomIDs = append(floors[i].Rooms[j].NextRoomIDs, strconv.Itoa(i+1)+","+strconv.Itoa(room_idx))
				}

				max_room_connect_idx = room_indexes_to_connect
				// 50% chance for better branching. hard to explain with just text, if you are reading this, ask me about it in the interview :D
				if max_room_connect_idx < nextFloorLen - 1 {
					max_room_connect_idx += rand.Intn(2) 
				}
			}
		}
	}
}
