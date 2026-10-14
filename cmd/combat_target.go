package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/objects"
)

// currentMobVictim is the mob the actor is already fighting, or nil.
func currentMobVictim(s *state) *objects.Mob {
	victim, _ := s.actor.Victim.(*objects.Mob)
	return victim
}

// hasMobTarget reports whether a targeted combat command has something to
// act on: a name in the input, or a mob the actor is already fighting.
func hasMobTarget(s *state) bool {
	return len(s.input) > 0 || currentMobVictim(s) != nil
}

// mobTarget finds the mob a targeted combat command acts on in this room:
// the one named in the input ("target #"), or with no name the mob the actor
// is already fighting. It is nil when that mob is not here.
func mobTarget(s *state) *objects.Mob {
	if len(s.input) < 1 {
		victim := currentMobVictim(s)
		if victim == nil {
			return nil
		}
		return s.where.Mobs.Search(victim.Name, s.where.Mobs.GetNumber(victim), s.actor)
	}
	nameNum := 1
	if len(s.words) > 1 {
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}
	return s.where.Mobs.Search(s.input[0], nameNum, s.actor)
}
