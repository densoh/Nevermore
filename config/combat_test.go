package config

import "testing"

func TestSpellTier(t *testing.T) {
	cases := map[int]int{1: 1, 4: 4, 7: 7, CombustMagnitude: 4, 0: 0, 8: 0}
	for magnitude, want := range cases {
		if got := SpellTier(magnitude); got != want {
			t.Errorf("SpellTier(%d) = %d, want %d", magnitude, got, want)
		}
	}
}

func TestIntMagicResist(t *testing.T) {
	cases := []struct{ intel, tier, want int }{
		{10, 7, 0},  // nothing at or below 10
		{5, 7, 0},   // and never negative
		{11, 1, 1},  // one point, plain
		{20, 0, 10}, // breath and other untiered hits take the plain amount
		{20, 2, 10}, // tiers 1-2 never drop below x1
		{20, 3, 15}, // x1.5
		{20, 4, 20}, // x2
		{20, 7, 35}, // x3.5
		{15, 3, 7},  // 5 * 1.5 rounds down
		{30, 6, 60}, // x3
	}
	for _, c := range cases {
		if got := IntMagicResist(c.intel, c.tier); got != c.want {
			t.Errorf("IntMagicResist(%d, %d) = %d, want %d", c.intel, c.tier, got, c.want)
		}
	}
}
