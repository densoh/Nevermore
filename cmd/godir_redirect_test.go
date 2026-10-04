package cmd

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

// A player walking through a temporarily redirected exit arrives in the
// temporary room, and in the exit's own room once the redirect is lifted.
func TestGoFollowsRedirectedExit(t *testing.T) {
	const from, home, away = -130, -131, -132
	room := func(id int) *objects.Room {
		r := &objects.Room{
			RoomId: id,
			Flags:  map[string]bool{"active": true},
			Exits:  map[string]*objects.Exit{},
			Chars:  &objects.CharInventory{ParentId: id},
			Mobs:   &objects.MobInventory{ParentId: id},
			Items:  objects.NewItemInventory(),
		}
		objects.Rooms[id] = r
		t.Cleanup(func() { delete(objects.Rooms, id) })
		return r
	}
	start, _, _ := room(from), room(home), room(away)
	bystander := func(id int) {
		objects.Rooms[id].Chars.Contents = append(objects.Rooms[id].Chars.Contents,
			&objects.Character{Object: objects.Object{Name: "Bystander"}, Flags: map[string]bool{}, ParentId: id})
	}
	bystander(home)
	bystander(away)
	start.Exits["gate"] = &objects.Exit{Object: objects.Object{Name: "gate", Placement: 3}, ParentId: from, ToId: home, Flags: map[string]bool{}}

	walker := func() *objects.Character {
		c := &objects.Character{
			Object:     objects.Object{Name: "Walker", Placement: 3},
			Permission: permissions.Anyone | permissions.Player,
			ParentId:   from,
			Flags:      map[string]bool{},
			Modifiers:  map[string]int{},
			Timers:     map[string]time.Time{},
			Hooks:      map[string]map[string]*objects.Hook{},
			Stam:       objects.Meter{Max: 10, Current: 10},
			Str:        objects.Meter{Max: 20, Current: 20},
			Inventory:  objects.NewItemInventory(),
			Equipment:  &objects.Equipment{},
		}
		// Someone stays behind, so leaving never empties the room and sets
		// off its last-person bookkeeping.
		start.Chars.Contents = []*objects.Character{c,
			{Object: objects.Object{Name: "Stayer"}, Flags: map[string]bool{}, ParentId: from}}
		return c
	}

	objects.RedirectExitFor(from, "gate", away, time.Hour)
	defer objects.RedirectExitFor(from, "gate", 0, 0)
	c := walker()
	Parse(c, "go gate")
	if c.ParentId != away {
		t.Errorf("with the redirect on, the player arrived in %d, want %d", c.ParentId, away)
	}

	objects.RedirectExitFor(from, "gate", 0, 0)
	c = walker()
	Parse(c, "go gate")
	if c.ParentId != home {
		t.Errorf("with the redirect lifted, the player arrived in %d, want %d", c.ParentId, home)
	}
}
