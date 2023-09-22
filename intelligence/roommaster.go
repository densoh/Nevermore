// Copyright 2023 Nevermore.

package intelligence

import (
	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"log"
	"time"
)

var ActiveRooms []int

func init() {
	objects.ActivateRoom = ActivateRoom
}

func ActivateRoom(roomId int) {
	log.Println("Adding room to active rooms: ", roomId)
	ActiveRooms = append(ActiveRooms, roomId)
	log.Println(ActiveRooms)
}

func DeactivateRoom(roomId int) {
	for c, p := range ActiveRooms {
		if p == roomId {
			copy(ActiveRooms[c:], ActiveRooms[c+1:])
			ActiveRooms = ActiveRooms[:len(ActiveRooms)-1]
			break
		}
	}
}

func StartRoomAI() {
	log.Println("Starting room AI")
	go func() {
		for {
			LoopRooms()
			time.Sleep(1 * time.Second)
		}
	}()
}

func LoopRooms() {
	for _, r := range ActiveRooms {
		objects.Rooms[r].Lock()
		//log.Println("Room AI invoked for room: ", r)
		if len(objects.Rooms[r].Chars.Contents) <= 0 {
			if !time.Time.IsZero(objects.Rooms[r].EvacuateTime) &&
				time.Now().Sub(objects.Rooms[r].EvacuateTime).Seconds() > float64(config.RoomClearTimer) {
				log.Println("Clear timer invoked for room: ", r)
				DeactivateRoom(r)
				objects.Rooms[r].CleanRoom()
			}
			return
		}
		if objects.Rooms[r].Flags["encounters_on"] {
			if time.Now().Sub(objects.Rooms[r].LastEncounterTime).Seconds() > float64(objects.Rooms[r].EncounterSpeed) {
				objects.Rooms[r].Encounter()
			}
		}

		if (objects.Rooms[r].Flags["fire"] ||
			objects.Rooms[r].Flags["earth"] ||
			objects.Rooms[r].Flags["wind"] ||
			objects.Rooms[r].Flags["water"]) &&
			time.Now().Sub(objects.Rooms[r].LastEffectTime).Seconds() > float64(config.RoomEffectInvocation) {
			objects.Rooms[r].ElementalDamage()
		}
		objects.Rooms[r].Unlock()
	}
}
