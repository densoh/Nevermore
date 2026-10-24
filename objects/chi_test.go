package objects

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
)

func monk(maxChi int) *Character {
	return &Character{
		Class: config.MONK,
		Flags: map[string]bool{},
		Mana:  Meter{Max: maxChi},
	}
}

// A single award never adds more than ChiGainCapPercent of the pool, so a
// multi-hit round on a small pool can't fill it in one swing.
func TestGainChiCappedPerAward(t *testing.T) {
	c := monk(100)
	cap := config.ChiGainCap(100)
	if got := c.GainChi(cap*5, false); got != cap {
		t.Errorf("oversized award added %d, want cap %d", got, cap)
	}
	if got := c.GainChi(cap-1, false); got != cap-1 {
		t.Errorf("award under the cap added %d, want %d", got, cap-1)
	}
}

// The cap applies after the meditate boost, so meditating can't push a
// capped award over the line.
func TestGainChiCapAppliesAfterMeditate(t *testing.T) {
	c := monk(100)
	c.Flags["meditate"] = true
	cap := config.ChiGainCap(100)
	if got := c.GainChi(cap, false); got != cap {
		t.Errorf("meditated award at the cap added %d, want %d", got, cap)
	}
}

// A tiny pool still earns at least one chi per award.
func TestGainChiCapFloorsAtOne(t *testing.T) {
	c := monk(4)
	if got := c.GainChi(3, false); got != 1 {
		t.Errorf("award on a 4-chi pool added %d, want 1", got)
	}
}

// Out-of-combat decay never takes chi below the tier floor (capped at 10):
// a bleed that would cross it is clamped, and one already at it is skipped.
func TestTickChiStopsAtFloor(t *testing.T) {
	c := monk(100)
	c.Tier = 15 // floor 10
	c.Mana.Current = 14
	c.tickChi() // LastCombat is zero, so the grace has long expired
	if c.Mana.Current != 10 {
		t.Fatalf("clamped decay left %d chi, want 10", c.Mana.Current)
	}
	c.tickChi()
	if c.Mana.Current != 10 {
		t.Fatalf("decay at the floor left %d chi, want 10", c.Mana.Current)
	}
	c.Tier = 4 // floor 4
	c.Mana.Current = 30
	c.tickChi()
	if c.Mana.Current != 20 {
		t.Fatalf("decay above the floor left %d chi, want 20", c.Mana.Current)
	}
}

// Out of combat a pool below the floor recovers up to it at the decay rate,
// never past it.
func TestTickChiRecoversToFloor(t *testing.T) {
	c := monk(50) // 5 chi per tick
	c.Tier = 15   // floor 10
	c.Mana.Current = 0
	c.tickChi()
	if c.Mana.Current != 5 {
		t.Fatalf("first recovery tick left %d chi, want 5", c.Mana.Current)
	}
	c.Mana.Current = 8
	c.tickChi()
	if c.Mana.Current != 10 {
		t.Fatalf("recovery past the floor left %d chi, want 10", c.Mana.Current)
	}
}

// Recovery waits for the same grace as decay and stops while meditating.
func TestTickChiRecoveryWaitsForGrace(t *testing.T) {
	c := monk(100)
	c.Tier = 10
	c.MarkCombat()
	c.tickChi()
	if c.Mana.Current != 0 {
		t.Fatalf("recovery inside the grace left %d chi, want 0", c.Mana.Current)
	}
	c.LastCombat = time.Time{}
	c.Flags["meditate"] = true
	c.tickChi()
	if c.Mana.Current != 0 {
		t.Fatalf("recovery while meditating left %d chi, want 0", c.Mana.Current)
	}
}

// wardMonk is a tier 10 monk with 20 con (a 14% ward), the given chi and
// plenty of stamina to take a hit.
func wardMonk(chi int) *Character {
	c := monk(100)
	c.Tier = 10
	c.Modifiers = map[string]int{}
	c.Con = Meter{Max: 20, Current: 20}
	c.Mana.Current = chi
	c.Stam = Meter{Max: 1000, Current: 1000}
	c.Vit = Meter{Max: 100, Current: 100}
	return c
}

// The ward takes its share of a magic hit and charges chi for it.
func TestChiWardAbsorbsMagic(t *testing.T) {
	c := wardMonk(50)
	stam, _, resisted := c.ReceiveMagicDamage(100, "fire", 0)
	if resisted != 14 || stam != 86 {
		t.Errorf("100 fire on a 14%% ward: took %d, resisted %d; want 86 and 14", stam, resisted)
	}
	if c.Mana.Current != 49 {
		t.Errorf("chi after turning aside 14 = %d, want 49", c.Mana.Current)
	}
}

// The ward works on what the other resists let through.
func TestChiWardAfterOtherResists(t *testing.T) {
	c := wardMonk(50)
	c.Flags["resist-fire"] = true
	stam, _, resisted := c.ReceiveMagicDamage(100, "fire", 0)
	// 20 resisted by the spell, then 14% of the remaining 80 is 11.
	if resisted != 31 || stam != 69 {
		t.Errorf("warded resist-fire monk took %d, resisted %d; want 69 and 31", stam, resisted)
	}
}

// Below tier 10, or with no chi, there is no ward.
func TestChiWardNeedsTierAndChi(t *testing.T) {
	low := wardMonk(50)
	low.Tier = 9
	if stam, _, _ := low.ReceiveMagicDamage(100, "fire", 0); stam != 100 {
		t.Errorf("tier 9 monk took %d, want the full 100", stam)
	}
	empty := wardMonk(0)
	if stam, _, _ := empty.ReceiveMagicDamage(100, "fire", 0); stam != 100 {
		t.Errorf("chi-less monk took %d, want the full 100", stam)
	}
}

// Int blocks a flat amount scaled by the spell's tier, and the chi ward
// works on what is left.
func TestIntResistThenChiWard(t *testing.T) {
	c := wardMonk(50)
	c.Int = Meter{Max: 20, Current: 20}
	// Tier 7: int blocks 10 * 3.5 = 35 of 350, then the 14% ward takes 44 of 315.
	stam, _, resisted := c.ReceiveMagicDamage(350, "fire", 7)
	if stam != 271 || resisted != 44 {
		t.Errorf("tier 7 hit on a 20 int monk: took %d, resisted %d; want 271 and 44", stam, resisted)
	}
}
