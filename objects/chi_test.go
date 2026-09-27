package objects

import (
	"testing"

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
