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
