package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $REQUIRE ITEM item_id [item_id...]
//
// Refuses the event unless its item is one of the listed item ids. The item
// stays with the player.
func init() {
	addHandler(scriptRequire{},
		"",
		permissions.Anyone,
		"$REQUIRE")
}

type scriptRequire cmd

func (scriptRequire) process(s *state) {
	ev := s.event
	if ev == nil {
		return
	}
	if len(s.words) < 2 || s.words[0] != "ITEM" {
		ev.refuse(s, "Nothing happens.")
		return
	}
	if ev.item != nil {
		for _, w := range s.words[1:] {
			if id, err := strconv.Atoi(w); err == nil && id == ev.item.ItemId {
				s.ok = true
				return
			}
		}
	}
	name := "that"
	if ev.item != nil {
		name = ev.item.Name
	}
	ev.refuse(s, "You can't put "+name+" into "+ev.owner+".")
}
