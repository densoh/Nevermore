package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
)

func testMob(name string) *objects.Mob {
	m := &objects.Mob{Flags: map[string]bool{}}
	m.Name = name
	return m
}

// targetState puts a monk in a room with two trolls and an orc.
func targetState(input ...string) (*state, []*objects.Mob) {
	mobs := []*objects.Mob{testMob("troll"), testMob("troll"), testMob("orc")}
	s := attackerState(config.MONK, nil)
	s.where = &objects.Room{Mobs: &objects.MobInventory{Contents: mobs}}
	s.input = input
	s.words = input
	return s, mobs
}

// A named target is found as usual, number included.
func TestMobTargetByName(t *testing.T) {
	s, mobs := targetState("troll", "2")
	if got := mobTarget(s); got != mobs[1] {
		t.Errorf("troll 2 resolved to %v, want the second troll", got)
	}
}

// With no name, the mob already being fought is used, even among namesakes.
func TestMobTargetFallsBackToVictim(t *testing.T) {
	s, mobs := targetState()
	s.actor.Victim = mobs[1]
	if !hasMobTarget(s) {
		t.Fatal("a monk fighting a troll has no target")
	}
	if got := mobTarget(s); got != mobs[1] {
		t.Errorf("no-name target resolved to %v, want the troll being fought", got)
	}
}

// No name and no fight: nothing to act on.
func TestMobTargetNoVictim(t *testing.T) {
	s, _ := targetState()
	if hasMobTarget(s) || mobTarget(s) != nil {
		t.Error("found a target with no name and no current fight")
	}
}

// A victim that has left the room is not swapped for a namesake.
func TestMobTargetVictimGone(t *testing.T) {
	s, _ := targetState()
	s.actor.Victim = testMob("troll")
	if got := mobTarget(s); got != nil {
		t.Errorf("departed victim resolved to %v, want nil", got)
	}
}
