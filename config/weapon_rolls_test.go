package config

import "testing"

func TestStandardWeaponRolls(t *testing.T) {
	realCap := RealWeaponCap(10, false) // 59
	if realCap != 59 {
		t.Fatalf("tier 10 real cap = %d, want 59", realCap)
	}
	rolls := StandardWeaponRolls(realCap)
	if len(rolls) == 0 {
		t.Fatal("no rolls for tier 10")
	}
	seen := map[WeaponRoll]bool{}
	for _, r := range rolls {
		if r.Max() < 55 || r.Max() > 57 {
			t.Errorf("%v max %d outside 55-57", r, r.Max())
		}
		if r.Min()*2 < realCap {
			t.Errorf("%v min %d under half the cap", r, r.Min())
		}
		if !StandardDiceShape(r.Dice, r.Sides) || r.Plus < 0 {
			t.Errorf("%v breaks the dice rules", r)
		}
		if seen[r] {
			t.Errorf("%v listed twice", r)
		}
		seen[r] = true
	}
	// The tier 10 template rolls from the consolidation pass.
	for _, want := range []WeaponRoll{{9, 3, 30}, {5, 5, 32}, {3, 7, 36}, {2, 11, 35}} {
		if !seen[want] {
			t.Errorf("%v missing", want)
		}
	}
}

func TestStandardDiceShape(t *testing.T) {
	for _, c := range []struct {
		dice, sides int
		want        bool
	}{
		{1, 20, true}, {1, 21, false}, {8, 2, true}, {9, 2, false},
		{10, 3, true}, {11, 3, false}, {12, 4, true}, {13, 4, false}, {3, 1, false},
	} {
		if got := StandardDiceShape(c.dice, c.sides); got != c.want {
			t.Errorf("%dd%d = %v, want %v", c.dice, c.sides, got, c.want)
		}
	}
}

func TestWeaponTier(t *testing.T) {
	for _, c := range []struct{ max, tier int }{{59, 10}, {60, 11}, {55, 10}, {54, 9}, {14, 1}, {500, 0}} {
		if got := WeaponTier(c.max, false); got != c.tier {
			t.Errorf("max %d = tier %d, want %d", c.max, got, c.tier)
		}
	}
	if got := WeaponTier(68, true); got != 10 {
		t.Errorf("2h max 68 = tier %d, want 10", got)
	}
}

func TestWeaponRollTargetPercent(t *testing.T) {
	for tier, want := range map[int]int{1: 78, 9: 78, 10: 80, 14: 80, 15: 82, 19: 82, 20: 84, 25: 84} {
		if got := WeaponRollTargetPercent(tier); got != want {
			t.Errorf("tier %d = %d, want %d", tier, got, want)
		}
	}
}
