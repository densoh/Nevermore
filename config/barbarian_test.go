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

func TestCrushingChanceIsDoubleCritical(t *testing.T) {
	if len(CrushingChance) != len(CriticalDamage) {
		t.Fatalf("table lengths differ: %d vs %d", len(CrushingChance), len(CriticalDamage))
	}
	for skill := range CriticalDamage {
		if CrushingChance[skill] != 2*CriticalDamage[skill] {
			t.Errorf("skill %d: crushing %d, want %d", skill, CrushingChance[skill], 2*CriticalDamage[skill])
		}
	}
}

func TestBashChancesCumulative(t *testing.T) {
	for skill, row := range BashChances {
		if len(row) != 3 {
			t.Errorf("skill %d: %d thresholds, want 3 (thunk, thwomp, thump)", skill, len(row))
			continue
		}
		if row[0] > row[1] || row[1] > row[2] {
			t.Errorf("skill %d: thresholds not cumulative: %v", skill, row)
		}
	}
}

func TestBashDamage(t *testing.T) {
	cases := []struct {
		hit  int
		mult float64
		tier int
		want int
	}{
		{66, 1, 10, 96},    // plain bash: hit + 3*tier
		{66, 10, 10, 690},  // thunk multiplies the hit, not the tier bonus
		{41, 1.5, 20, 122}, // thump rounds the hit up before the bonus
	}
	for _, c := range cases {
		if got := BashDamage(c.hit, c.mult, c.tier); got != c.want {
			t.Errorf("BashDamage(%d, %v, %d) = %d, want %d", c.hit, c.mult, c.tier, got, c.want)
		}
	}
}

func TestRollBashUnknownSkillIsPlain(t *testing.T) {
	for _, skill := range []int{-1, 10, 42} {
		dmg, stun, msg := RollBash(skill, 10)
		if dmg != 1 || stun != 1 || msg != "" {
			t.Errorf("RollBash(%d) = %v, %d, %q; want a plain bash", skill, dmg, stun, msg)
		}
	}
}

func TestThunkMultiplier(t *testing.T) {
	cases := []struct {
		tier int
		want float64
	}{
		{1, 5}, {5, 5}, // flat through tier 5
		{6, 5.5}, {15, 10}, {25, 15}, // +0.5 per tier after
	}
	for _, c := range cases {
		if got := ThunkMultiplier(c.tier); got != c.want {
			t.Errorf("ThunkMultiplier(%d) = %v, want %v", c.tier, got, c.want)
		}
	}
}
