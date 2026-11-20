package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
)

func sneakThief(tier int, weaponType int) *objects.Character {
	c := &objects.Character{
		Class:     config.THIEF,
		Tier:      tier,
		Equipment: &objects.Equipment{Main: &objects.Item{ItemType: weaponType}},
	}
	c.Name = "Sly"
	c.ParentId = -9201
	return c
}

func sneakMob(target string) *objects.Mob {
	m := &objects.Mob{CurrentTarget: target, Stam: objects.Meter{Max: 100, Current: 100}}
	return m
}

func TestSneakAttackOdds(t *testing.T) {
	thief := sneakThief(12, 0)
	hit := sneakMob("Tank")
	room := &objects.Room{Mobs: &objects.MobInventory{Contents: []*objects.Mob{hit}}}
	objects.Rooms[thief.ParentId] = room
	defer delete(objects.Rooms, thief.ParentId)

	if ok, chance := sneakAttackOdds(thief, hit); !ok || chance != 100 {
		t.Errorf("nothing on the thief: ok=%v chance=%d, want true 100", ok, chance)
	}

	// Two other mobs on the thief, plus a dead one that doesn't count.
	dead := sneakMob("Sly")
	dead.Stam.Current = 0
	room.Mobs.Contents = append(room.Mobs.Contents, sneakMob("Sly"), sneakMob("Sly"), dead)
	if ok, chance := sneakAttackOdds(thief, hit); !ok || chance != 50 {
		t.Errorf("two others on the thief: ok=%v chance=%d, want true 50", ok, chance)
	}

	// The hit mob targeting the thief: never.
	if ok, _ := sneakAttackOdds(thief, room.Mobs.Contents[1]); ok {
		t.Error("mob targeting the thief allowed a sneak attack")
	}
}

func TestSneakAttackOddsGates(t *testing.T) {
	hit := sneakMob("Tank")
	cases := map[string]*objects.Character{
		"below tier 10": sneakThief(9, 0),
		"ranged weapon": sneakThief(15, 4),
	}
	fighter := sneakThief(15, 0)
	fighter.Class = config.FIGHTER
	cases["not a thief"] = fighter
	unarmed := sneakThief(15, 0)
	unarmed.Equipment.Main = nil
	cases["no weapon"] = unarmed
	for name, c := range cases {
		if ok, _ := sneakAttackOdds(c, hit); ok {
			t.Errorf("%s: sneak attack allowed", name)
		}
	}
}
