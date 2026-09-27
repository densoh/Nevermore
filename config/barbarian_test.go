package config

import "testing"

func TestBerserkDamageBonus(t *testing.T) {
	cases := []struct{ str, tier, want int }{
		{25, 7, 89},
		{25, 20, 115},
		{45, 20, 175},
	}
	for _, c := range cases {
		if got := BerserkDamageBonus(c.str, c.tier); got != c.want {
			t.Errorf("BerserkDamageBonus(%d, %d) = %d, want %d", c.str, c.tier, got, c.want)
		}
	}
}

func TestRecklessStamCost(t *testing.T) {
	cases := map[int]int{10: 1, 154: 8, 220: 11, 440: 22}
	for maxStam, want := range cases {
		if got := RecklessStamCost(maxStam); got != want {
			t.Errorf("RecklessStamCost(%d) = %d, want %d", maxStam, got, want)
		}
	}
}

func TestBerserkDuration(t *testing.T) {
	if got := BerserkDurationFor(10); got != 70 {
		t.Errorf("BerserkDurationFor(10) = %d, want 70", got)
	}
	if got := BerserkMaxDurationFor(20); got != 140 {
		t.Errorf("BerserkMaxDurationFor(20) = %d, want 140", got)
	}
}
