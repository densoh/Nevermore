package objects

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
)

// Expose takes ExposeWeaknessArmor off once; a room reset (Rest) clears it.
func TestExposeAndReset(t *testing.T) {
	m := &Mob{Armor: 60, Stam: Meter{Max: 100, Current: 100}, Flags: map[string]bool{}}
	if !m.Expose() || m.Expose() {
		t.Fatal("expose should land once and refuse a second time")
	}
	if got := m.EffectiveArmor(); got != 60-config.ExposeWeaknessArmor {
		t.Fatalf("exposed armor %d", got)
	}
	// Armor 40 blocks 20% of a hit.
	if dealt, _, _ := m.ReceiveDamage(100); dealt != 80 {
		t.Errorf("exposed mob took %d of 100, want 80", dealt)
	}
	m.StoppedAt = time.Now().Add(-time.Minute)
	m.Rest()
	if m.CheckFlag("exposed") || m.EffectiveArmor() != 60 {
		t.Error("room reset left the mob exposed")
	}
}

// Poison ticks take only poison resistance, ignore armor, and train the
// owner's stealth.
func TestPoisonStrikeTick(t *testing.T) {
	const roomId = -4343
	owner := &Character{Object: Object{Name: "Sly"}, Skills: map[int]*Accumulator{11: {}}}
	mob := &Mob{
		Object:           Object{Name: "Target"},
		Level:            10,
		ExpMultiplier:    1,
		Armor:            100,
		PoisonResistance: 100,
		Stam:             Meter{Max: 1000, Current: 1000},
		Flags:            map[string]bool{},
		ThreatTable:      map[string]int{},
	}
	mob.ParentId = roomId
	Rooms[roomId] = &Room{RoomId: roomId, Mobs: &MobInventory{Contents: []*Mob{mob}}, Chars: &CharInventory{Contents: []*Character{owner}}}
	defer delete(Rooms, roomId)

	ctx := &HitContext{CasterName: "Sly", Tier: 16, Element: "poison", StealthRank: 5}
	dotTick(poisonStrike, ctx, mob, 1)
	if mob.Stam.Current != 1000 {
		t.Fatalf("100%% poison resist still took %d", 1000-mob.Stam.Current)
	}

	mob.PoisonResistance = 0
	dotTick(poisonStrike, ctx, mob, 1)
	took := 1000 - mob.Stam.Current
	if took < 3 || took > 12 { // 3d4, armor ignored
		t.Errorf("tick took %d, want 3d4", took)
	}
	if owner.Skills[11].Value <= 0 {
		t.Error("poison tick didn't train stealth")
	}
}

// The fight ending drops the queued strike of a thief who fought the mob;
// one who queued before the pull keeps theirs.
func TestDeviousDroppedWhenFightEnds(t *testing.T) {
	const roomId = -4344
	fighter := &Character{Object: Object{Name: "Sly"}, DeviousStrike: config.DeviousPoison}
	waiting := &Character{Object: Object{Name: "Patient"}, DeviousStrike: config.DeviousExpose}
	mob := &Mob{Stam: Meter{Max: 10, Current: 0}, ThreatTable: map[string]int{"Sly": 5}}
	other := &Mob{Stam: Meter{Max: 10, Current: 10}, ThreatTable: map[string]int{}}
	inv := &MobInventory{ParentId: roomId, Contents: []*Mob{mob, other}}
	Rooms[roomId] = &Room{RoomId: roomId, Mobs: inv, Chars: &CharInventory{Contents: []*Character{fighter, waiting}}}
	defer delete(Rooms, roomId)

	inv.Remove(mob)
	if fighter.DeviousStrike != "" {
		t.Error("strike kept after the fight ended")
	}
	if waiting.DeviousStrike != config.DeviousExpose {
		t.Error("a thief who never fought lost their strike")
	}
}

// Leaving a room lets the readied strike go.
func TestDeviousDroppedOnLeavingRoom(t *testing.T) {
	const roomId = -4345
	thief := &Character{Object: Object{Name: "Sly"}, DeviousStrike: config.DeviousPoison}
	stays := &Character{Object: Object{Name: "Tank"}}
	inv := &CharInventory{ParentId: roomId, Contents: []*Character{thief, stays}}
	Rooms[roomId] = &Room{RoomId: roomId, Chars: inv, Mobs: &MobInventory{ParentId: roomId}}
	defer delete(Rooms, roomId)

	inv.Remove(thief)
	if thief.DeviousStrike != "" {
		t.Error("strike kept after leaving the room")
	}
}
