package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func TestSpellMinTier(t *testing.T) {
	cases := map[string]int{"vigor": 1, "mend": 3, "detraumatize": 7, "renewal": 11}
	for spell, want := range cases {
		if got := spellMinTier(spell); got != want {
			t.Errorf("spellMinTier(%q) = %d, want %d", spell, got, want)
		}
	}
}

// A device heals as a caster of the spell's lowest casting tier with
// BaseDevicePiety, so its range follows the spell's own formula.
func TestDeviceHealRanges(t *testing.T) {
	pie := config.BaseDevicePiety
	cases := []struct {
		spell    string
		roll     healRoll
		min, max float64
	}{
		// 8*.6 + (4..8) + 1/4
		{"vigor", minorHealRoll, 9.05, 13.05},
		// 8*.6 + (4..8) + 3/4
		{"mend", minorHealRoll, 9.55, 13.55},
		// 15 + 7/2 + 8*.7 + (1..10)
		{"detraumatize", detraumatizeRoll, 25.1, 34.1},
		// 30 + 11 + 8*.8 + 2d10
		{"renewal", renewalRoll, 49.4, 67.4},
	}
	for _, tc := range cases {
		tier := spellMinTier(tc.spell)
		for i := 0; i < 2000; i++ {
			got := tc.roll(tier, pie)
			if got < tc.min-1e-9 || got > tc.max+1e-9 {
				t.Fatalf("%s device roll %v outside [%v, %v]", tc.spell, got, tc.min, tc.max)
			}
		}
	}
}
