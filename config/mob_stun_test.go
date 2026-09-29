package config

import "testing"

func TestMobStunDuration(t *testing.T) {
	cases := []struct {
		mobLevel, playerTier, want int
	}{
		{10, 10, MobStunBase},     // equal level: full duration
		{15, 10, MobStunBase},     // higher-level mob: full duration
		{8, 10, MobStunBase - 2},  // mob two levels lower
		{1, 12, MobStunBase - 11}, // mob eleven levels lower
		{1, 25, MobStunMin},       // floors at the minimum
	}
	for _, c := range cases {
		if got := MobStunDuration(c.mobLevel, c.playerTier); got != c.want {
			t.Errorf("MobStunDuration(%d, %d) = %d, want %d", c.mobLevel, c.playerTier, got, c.want)
		}
	}
}

func TestStunSpellDuration(t *testing.T) {
	cases := []struct {
		intel, want int
	}{
		{0, 12},
		{11, 14},
		{12, 15}, // matches the old flat 15s
		{20, 17},
		{30, 19},
	}
	for _, c := range cases {
		if got := StunSpellDuration(c.intel); got != c.want {
			t.Errorf("StunSpellDuration(%d) = %d, want %d", c.intel, got, c.want)
		}
	}
}

func TestStunSpellSuccessChance(t *testing.T) {
	cases := []struct {
		tier, mobLevel, intel, mobInt, want int
	}{
		{10, 10, 40, 0, 90},  // same level, 40 int vs 0
		{10, 10, 20, 20, 70}, // even match
		{12, 10, 20, 20, 90}, // two levels over
		{10, 12, 20, 20, 50}, // two levels under
		{10, 10, 15, 20, 68}, // int deficit rounds toward zero
	}
	for _, c := range cases {
		if got := StunSpellSuccessChance(c.tier, c.mobLevel, c.intel, c.mobInt); got != c.want {
			t.Errorf("StunSpellSuccessChance(%d, %d, %d, %d) = %d, want %d", c.tier, c.mobLevel, c.intel, c.mobInt, got, c.want)
		}
	}
}
