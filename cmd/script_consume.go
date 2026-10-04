package cmd

import (
	"log"

	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $CONSUME
//
// Destroys the event's item. Put it after any step that can refuse, or the
// item is gone even though the event was turned down.
func init() {
	addHandler(scriptConsume{},
		"",
		permissions.Anyone,
		"$CONSUME")
}

type scriptConsume cmd

func (scriptConsume) process(s *state) {
	ev := s.event
	if ev == nil || ev.item == nil || ev.consumed {
		return
	}
	if err := s.actor.Inventory.Remove(ev.item); err != nil {
		log.Println("Error consuming item from inventory: ", err)
		ev.refuse(s, "Nothing happens.")
		return
	}
	ev.consumed = true
	s.ok = true
}
