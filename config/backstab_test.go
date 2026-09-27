package config

import "testing"

func TestBackstabMissChance(t *testing.T) {
	// The inverse of the old hit formula: 20% base plus 3 per stealth level.
	if got := BackstabMissChance(0); got != 80 {
		t.Errorf("stealth 0 miss = %d, want 80", got)
	}
	if got := BackstabMissChance(10); got != 50 {
		t.Errorf("stealth 10 miss = %d, want 50", got)
	}
}
