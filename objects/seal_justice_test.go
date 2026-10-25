package objects

import (
	"math"
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func justicePaladin(tier int, pie int, weaponLevel int) *Character {
	return &Character{
		Class:     config.PALADIN,
		Tier:      tier,
		Flags:     map[string]bool{},
		Modifiers: map[string]int{},
		Pie:       Meter{Current: pie},
		Equipment: &Equipment{Main: &Item{ItemType: 1}},
		Skills:    map[int]*Accumulator{1: {Value: config.SkillExpLevels[weaponLevel]}},
	}
}

func TestSealJusticeBonus(t *testing.T) {
	cases := []struct {
		tier, pie, weaponLevel int
		want                   float64
	}{
		{10, 13, 4, 3 + 13.0/3 + 2}, // tier 10 on the realistic skill schedule
		{15, 20, 6, 8 + 20.0/3 + 3}, // tier 15
		{20, 30, 9, 13 + 10 + 4.5},  // tier 20, paladins stop at weapon level 9
		{20, 30, 10, 13 + 10 + 4.5}, // grandmaster exp still caps at level 9
		{20, 45, 9, 13 + 15 + 4.5},  // gnome piety cap
		{5, 15, 0, 5},               // the tier term never goes negative
	}
	for _, tc := range cases {
		got := justicePaladin(tc.tier, tc.pie, tc.weaponLevel).SealJusticeBonus() * 100
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("tier %d pie %d level %d = %.2f%%, want %.2f%%", tc.tier, tc.pie, tc.weaponLevel, got, tc.want)
		}
	}
}
