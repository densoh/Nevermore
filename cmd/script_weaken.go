package cmd

import (
	"log"
	"strconv"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
)

// Syntax: $WEAKEN room_id mob_id percent max_steps
//
// Takes percent% of a mob's original armor, resistances, max stamina and
// damage off its template, up to max_steps times; see Mob.Weaken. The change
// is saved, so it is permanent and every later spawn has it. Copies of the
// mob standing in room_id pick it up straight away. Refuses once max_steps
// have been reached.
func init() {
	addHandler(scriptWeaken{},
		"",
		permissions.Anyone,
		"$WEAKEN")
}

type scriptWeaken cmd

func (scriptWeaken) process(s *state) {
	refuse := func(why string) {
		log.Println("$WEAKEN " + why)
		if s.event != nil {
			s.event.refuse(s, "Nothing happens.")
		}
	}

	if len(s.words) < 4 {
		refuse("needs room_id mob_id percent max_steps")
		return
	}
	var args [4]int
	for i := range args {
		v, err := strconv.Atoi(s.words[i])
		if err != nil {
			refuse("argument " + s.words[i] + " is not a number")
			return
		}
		args[i] = v
	}
	roomId, mobId, pct, maxSteps := args[0], args[1], args[2], args[3]

	room, ok := objects.Rooms[roomId]
	if !ok {
		refuse("room " + s.words[0] + " does not exist")
		return
	}
	template, ok := objects.Mobs[mobId]
	if !ok {
		refuse("mob " + s.words[1] + " does not exist")
		return
	}
	if !objects.ValidWeaken(pct, maxSteps) {
		refuse("percent times max_steps must be under 100")
		return
	}
	// lockScriptRooms took this lock before the script ran; if it isn't
	// held this step came from somewhere that skipped it.
	if !utils.IntIn(roomId, s.rLocks) {
		refuse("ran without the lock for room " + s.words[0])
		return
	}

	// At the cap is the normal end of the quest, not an error.
	if !template.Weaken(pct, maxSteps) {
		if s.event != nil {
			s.event.refuse(s, "Nothing happens.")
		}
		return
	}
	template.Save()

	for _, mob := range room.Mobs.Contents {
		if mob.MobId == mobId {
			mob.SyncWeakened(template)
		}
	}
	s.ok = true
}
