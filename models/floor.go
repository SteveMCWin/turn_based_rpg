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
	Id          string
	Encounter   Encounter
	NextRoomIDs []string
	IsCompleted bool
	CanEnter    bool
	Environment Environment
}

type Environment struct {
	Id          string
	Name        string
	Description string
}

type Floor struct {
	Rooms       []Room
	IsCompleted bool
	Idx         int
}

func MakeFirstRealm(numFloors, maxRoomsPerFloor, monster_spawn_chance int) []Floor {

	floors := make([]Floor, 1)

	floors[0] = Floor{
		Rooms: []Room{Room{
			Id:        "0" + ROOM_ID_SEPARATOR + "0",
			CanEnter:  true,
			Encounter: Encounter{Kind: EncounterKindMonster},
		}},
		IsCompleted: false,
		Idx:         0,
	}

	rest_of_the_floors := GenerateFloors(numFloors-1, maxRoomsPerFloor, 1, monster_spawn_chance)

	floors = append(floors, rest_of_the_floors...)

	return floors
}

func AddRealmToExistingOne(existing []Floor, numFloors, maxRoomsPerFloor, monster_spawn_chance int, monsterTemplates, bossTemplates []Monster, eventTemplates []Event, environmentTemplates []Environment) []Floor {

	new_realm := GenerateFloors(numFloors, maxRoomsPerFloor, len(existing), monster_spawn_chance)

	FillFloorEncounters(new_realm, monsterTemplates, bossTemplates, eventTemplates, environmentTemplates)
	ConnectFloorRooms(&existing[len(existing)-1], &new_realm[0])

	// CompleteRoom already ran for the boss before AddRealm was called, so
	// the new entry rooms were never enabled by it. Enable them now.
	for j := range new_realm[0].Rooms {
		new_realm[0].Rooms[j].CanEnter = true
	}

	existing = append(existing, new_realm...)

	return existing
}

func GenerateFloors(numFloors, maxRoomsPerFloor, start_floor_idx, monster_spawn_chance int) []Floor {
	floors := make([]Floor, numFloors)

	bossFloorIdx := start_floor_idx + numFloors - 1
	floors[numFloors-1] = Floor{
		Rooms: []Room{Room{
			Id:       strconv.Itoa(bossFloorIdx) + ROOM_ID_SEPARATOR + "0",
			Encounter: Encounter{Kind: EncounterKindBoss},
		}},
		IsCompleted: false,
		Idx:         bossFloorIdx,
	}

	for i := 0; i < numFloors-1; i++ {
		floors[i].Idx = start_floor_idx + i
		floors[i].Rooms = make([]Room, rand.Intn(maxRoomsPerFloor)+1)
		for j := range floors[i].Rooms {
			floors[i].Rooms[j].Id = strconv.Itoa(floors[i].Idx) + ROOM_ID_SEPARATOR + strconv.Itoa(j)

			is_monster_room := rand.Intn(100) < monster_spawn_chance
			if is_monster_room {
				floors[i].Rooms[j].Encounter.Kind = EncounterKindMonster
			} else {
				floors[i].Rooms[j].Encounter.Kind = EncounterKindEvent
			}
		}
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

func ConnectFloorRooms(f1, f2 *Floor) {
	last_conn_idx := 0
	f1_len := len(f1.Rooms)
	f2_len := len(f2.Rooms)

	for i := range f1_len {
		var room_indexes_to_connect int

		isLastRoom := i == f1_len-1

		if isLastRoom {
			// Ensure all next-floor rooms up to the end are reachable
			room_indexes_to_connect = f2_len - 1
		} else {
			room_indexes_to_connect = max(last_conn_idx, rand.Intn(f2_len))
		}

		for room_idx := last_conn_idx; room_idx <= room_indexes_to_connect; room_idx++ {
			f1.Rooms[i].NextRoomIDs = append(f1.Rooms[i].NextRoomIDs, strconv.Itoa(f2.Idx)+","+strconv.Itoa(room_idx))
		}

		last_conn_idx = room_indexes_to_connect
		// 50% chance for better branching. hard to explain with just text, if you are reading this, ask me about it in the interview :D
		if last_conn_idx < f2_len-1 {
			last_conn_idx += rand.Intn(2)
		}
	}

}

func FillFloorEncounters(floors []Floor, monsterTemplates, bossTemplates []Monster, eventTemplates []Event, environmentTemplates []Environment) {
	events := slices.Clone(eventTemplates)
	monsters := slices.Clone(monsterTemplates)
	bosses := slices.Clone(bossTemplates)
	environments := slices.Clone(environmentTemplates)

	rand_monster_indexes := rand.Perm(len(monsters))
	rand_boss_indexes := rand.Perm(len(bosses))
	rand_event_indexes := rand.Perm(len(events))

	for i := range floors {
		for j := range floors[i].Rooms {
			switch floors[i].Rooms[j].Encounter.Kind {
			case EncounterKindBoss:
				floors[i].Rooms[j].Encounter.Monster = &bosses[rand_boss_indexes[len(rand_boss_indexes)-1]]
				floors[i].Rooms[j].Environment = environments[rand.Intn(len(environments))]
				rand_boss_indexes = rand_boss_indexes[:len(rand_boss_indexes)-1]
				if len(rand_boss_indexes) <= 0 {
					rand_boss_indexes = rand.Perm(len(bosses))
				}
				floors[i].Rooms[j].Encounter.Monster.SetToLevel(floors[i].Idx + 1)
			case EncounterKindMonster:
				floors[i].Rooms[j].Encounter.Monster = &monsters[rand_monster_indexes[len(rand_monster_indexes)-1]]
				floors[i].Rooms[j].Environment = environments[rand.Intn(len(environments))]
				rand_monster_indexes = rand_monster_indexes[:len(rand_monster_indexes)-1]
				if len(rand_monster_indexes) <= 0 {
					rand_monster_indexes = rand.Perm(len(monsters))
				}
				floors[i].Rooms[j].Encounter.Monster.SetToLevel(floors[i].Idx + 1)
			default:
				floors[i].Rooms[j].Encounter.Event = &events[rand_event_indexes[len(rand_event_indexes)-1]]
				rand_event_indexes = rand_event_indexes[:len(rand_event_indexes)-1]
				if len(rand_event_indexes) <= 0 {
					rand_event_indexes = rand.Perm(len(events))
				}
			}
		}

		if i < len(floors)-1 {
			ConnectFloorRooms(&floors[i], &floors[i+1])
		}
	}
}
