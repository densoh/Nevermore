package config

import "testing"

func TestBardSingCastMod(t *testing.T) {
	if got := BardSingCastMod(BARD, 1, true); got != .8 {
		t.Errorf("tier 1 bard singing = %v, want 0.8", got)
	}
	if got := BardSingCastMod(BARD, BardSingCastFreeTier-1, true); got != .8 {
		t.Errorf("tier %d bard singing = %v, want 0.8", BardSingCastFreeTier-1, got)
	}
	if got := BardSingCastMod(BARD, BardSingCastFreeTier, true); got != 1 {
		t.Errorf("tier %d bard singing = %v, want 1", BardSingCastFreeTier, got)
	}
	if got := BardSingCastMod(BARD, 1, false); got != 1 {
		t.Errorf("bard not singing = %v, want 1", got)
	}
	if got := BardSingCastMod(CLERIC, 1, true); got != 1 {
		t.Errorf("non-bard = %v, want 1", got)
	}
}
