package objects

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTriggerOnHitChance(t *testing.T) {
	calls := 0
	HitCallbacks["test"] = func(*HitContext, *Mob) { calls++ }
	defer delete(HitCallbacks, "test")

	mob := &Mob{Stam: Meter{Max: 100, Current: 100}}
	TriggerOnHit([]OnHit{{Name: "test", Chance: 0}}, &HitContext{}, mob)
	if calls != 0 {
		t.Fatalf("0%% callback fired %d times", calls)
	}
	TriggerOnHit([]OnHit{{Name: "test", Chance: 100}, {Name: "missing", Chance: 100}}, &HitContext{}, mob)
	if calls != 1 {
		t.Fatalf("100%% callback fired %d times, want 1", calls)
	}
	mob.Stam.Current = 0
	TriggerOnHit([]OnHit{{Name: "test", Chance: 100}}, &HitContext{}, mob)
	if calls != 1 {
		t.Fatal("callback fired on a dead mob")
	}
}

func TestBurnDamageScalesWithBonus(t *testing.T) {
	// 1d1 always rolls 1.
	if got := burnDamage(&HitContext{Tier: 1, Bonus: 2.5}); got != 2 {
		t.Fatalf("burn damage %d, want 2", got)
	}
	for i := 0; i < 200; i++ {
		if got := burnDamage(&HitContext{Tier: 14, Bonus: 1}); got < 1 || got > 14 {
			t.Fatalf("1d14 burn rolled %d", got)
		}
	}
}

func TestBurnCredit(t *testing.T) {
	caster := &Character{Object: Object{Name: "Mage"}}
	tank := &Character{Object: Object{Name: "Tank"}}
	thief := &Character{Object: Object{Name: "Thief"}}
	mob := &Mob{ThreatTable: map[string]int{"Mage": 50, "Tank": 30, "Thief": 10}}

	room := &Room{Chars: &CharInventory{Contents: []*Character{thief, caster, tank}}}
	if got := burnCredit("Mage", mob, room); got != caster {
		t.Errorf("caster in room: credited %v", got)
	}
	room.Chars.Contents = []*Character{thief, tank}
	if got := burnCredit("Mage", mob, room); got != tank {
		t.Errorf("caster gone: credited %v, want top threat Tank", got)
	}
	room.Chars.Contents = nil
	if got := burnCredit("Mage", mob, room); got != nil {
		t.Errorf("empty room: credited %v, want nobody", got)
	}
}

func TestMobEffectTicksThenExpires(t *testing.T) {
	const roomId = -4242
	mob := &Mob{
		Object:  Object{Name: "Target"},
		Stam:    Meter{Max: 100, Current: 100},
		Effects: map[string]*Effect{},
		Flags:   map[string]bool{},
	}
	mob.ParentId = roomId
	// Left registered: the effect's timers can outlive the test.
	Rooms[roomId] = &Room{RoomId: roomId, Mobs: &MobInventory{Contents: []*Mob{mob}}, Chars: &CharInventory{}}

	var ticks, offs atomic.Int32
	Rooms[roomId].LockRoom("test", true)
	mob.ApplyEffect("dot", "3", 1, 0,
		func(triggers int) { ticks.Add(1) },
		func() { offs.Add(1) })
	Rooms[roomId].UnlockRoom("test", true)

	time.Sleep(4500 * time.Millisecond)
	Rooms[roomId].LockRoom("test", true)
	_, still := mob.Effects["dot"]
	Rooms[roomId].UnlockRoom("test", true)

	// One run on apply plus three timed ticks.
	if got := ticks.Load(); got != 4 {
		t.Errorf("effect ran %d times, want 4", got)
	}
	if offs.Load() != 1 || still {
		t.Errorf("effect did not expire cleanly: effectOff ran %d times, still applied %v", offs.Load(), still)
	}
}

func TestMobReplaceEffectStopsOldTicks(t *testing.T) {
	const roomId = -4243
	mob := &Mob{
		Object:  Object{Name: "Target"},
		Stam:    Meter{Max: 100, Current: 100},
		Effects: map[string]*Effect{},
		Flags:   map[string]bool{},
	}
	mob.ParentId = roomId
	// Left registered: the effect's timers can outlive the test.
	Rooms[roomId] = &Room{RoomId: roomId, Mobs: &MobInventory{Contents: []*Mob{mob}}, Chars: &CharInventory{}}

	var oldTicks atomic.Int32
	Rooms[roomId].LockRoom("test", true)
	mob.ApplyEffect("dot", "10", 1, 0, func(int) { oldTicks.Add(1) }, func() {})
	mob.ReplaceEffect("dot", "10", 5, 0, func(int) {}, func() {})
	Rooms[roomId].UnlockRoom("test", true)

	time.Sleep(1500 * time.Millisecond)
	if got := oldTicks.Load(); got != 1 {
		t.Errorf("replaced effect kept ticking: ran %d times, want only the initial run", got)
	}
}
