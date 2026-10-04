package cmd

import (
	"log"
	"strings"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $EXITTO room_id exit to_room_id seconds
//
// Makes an exit lead to another room for a while. The exit itself is not
// changed: when the time is up, or after a reboot, it leads where it always
// did. Seconds of 0 puts it back at once.
func init() {
	addHandler(scriptExitTo{},
		"",
		permissions.Anyone,
		"$EXITTO")
}

type scriptExitTo cmd

func (scriptExitTo) process(s *state) {
	roomId, exit, to, d, ok := objects.ParseExitTo(s.words)
	if !ok {
		log.Println("$EXITTO couldn't use: " + strings.Join(s.words, " "))
		if s.event != nil {
			s.event.refuse(s, "Nothing happens.")
		}
		return
	}
	objects.RedirectExitFor(roomId, exit, to, d)
	s.ok = true
}
