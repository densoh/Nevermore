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
		// pie 8 is under the floor, so (8..13) + 1/4
		{"vigor", minorHealRoll, 8.25, 13.25},
		// (8..13) + 3/4
		{"mend", minorHealRoll, 8.75, 13.75},
		// 18 + 7/3 + 8*.7 + (1..10)
		{"detraumatize", detraumatizeRoll, 18 + 7.0/3 + 5.6 + 1, 18 + 7.0/3 + 5.6 + 10},
		// 35 + 11/2 + 8*.8 + 2d8
		{"renewal", renewalRoll, 48.9, 62.9},
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

// Vigor/mend count only piety above the floor, and spellcasting classes gain
// the tier bonus at tier/3 instead of tier/4.
func TestMinorHealShape(t *testing.T) {
	cases := []struct {
		name     string
		roll     healRoll
		tier     int
		pie      float64
		min, max float64
	}{
		{"low pie counts nothing", minorHealRoll, 12, 8, 8 + 3, 13 + 3},
		{"pie above floor", minorHealRoll, 12, 20, 9 + 8 + 3, 9 + 13 + 3},
		{"caster tier bonus", casterMinorHealRoll, 12, 20, 9 + 8 + 4, 9 + 13 + 4},
	}
	for _, tc := range cases {
		for i := 0; i < 2000; i++ {
			got := tc.roll(tc.tier, tc.pie)
			if got < tc.min-1e-9 || got > tc.max+1e-9 {
				t.Fatalf("%s: roll %v outside [%v, %v]", tc.name, got, tc.min, tc.max)
			}
		}
	}
}
