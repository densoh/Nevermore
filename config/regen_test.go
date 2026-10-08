package config

import "testing"

func TestRegenStatUsesIntAverageWhenHigher(t *testing.T) {
	if got := RegenStat(25, 10); got != 25 {
		t.Errorf("pie-heavy: got %v, want 25", got)
	}
	if got := RegenStat(10, 30); got != 20 {
		t.Errorf("int-heavy: got %v, want 20", got)
	}
	if got := RegenStat(15, 25); got != 20 {
		t.Errorf("balanced: got %v, want 20", got)
	}
}

func TestManaRegenFormula(t *testing.T) {
	cases := []struct {
		pie, intel, tier int
		inCombat         bool
		want             int
	}{
		{20, 10, 1, false, 7},   // ceil(0.3*21)
		{20, 10, 9, false, 9},   // ceil(0.3*29)
		{25, 10, 25, false, 15}, // ceil(0.3*50)
		{15, 25, 16, false, 11}, // avg 20 + 16 = 36 -> ceil(10.8)
		{15, 25, 16, true, 4},   // cut before rounding: ceil(10.8*.35 = 3.78)
		{25, 10, 25, true, 6},   // ceil(15*.35 = 5.25)
		{10, 30, 10, false, 9},  // int-heavy mage: avg 20 + 10 -> 9
	}
	for _, c := range cases {
		if got := ManaRegen(c.pie, c.intel, c.tier, 1, c.inCombat, false); got != c.want {
			t.Errorf("ManaRegen(%d,%d,%d,combat=%v) = %d, want %d", c.pie, c.intel, c.tier, c.inCombat, got, c.want)
		}
	}
}

func TestManaRegenRoomAndCombatStack(t *testing.T) {
	// heal_fast doubles before the combat cut: 9*2*.35 = 6.3 -> 7.
	base := ManaRegen(20, 10, 10, 1, false, false)
	if got := ManaRegen(20, 10, 10, HealFastRoomRegenMod, true, false); got != 7 {
		t.Errorf("heal_fast + combat = %d, want 7", got)
	}
	if got := ManaRegen(20, 10, 10, HealFastRoomRegenMod, false, false); got != 2*base {
		t.Errorf("heal_fast = %d, want %d", got, 2*base)
	}
}

func TestHealthRegenCutInCombat(t *testing.T) {
	if got := HealthRegen(20, 1, false, false); got != 2 {
		t.Errorf("con 20 = %d, want 2", got)
	}
	if got := HealthRegen(20, 1, true, false); got != 1 {
		t.Errorf("con 20 in combat = %d, want 1", got)
	}
	// Rounding happens after the cut, so a tiny regen still yields 1.
	if got := HealthRegen(5, 1, true, false); got != 1 {
		t.Errorf("con 5 in combat = %d, want 1", got)
	}
}

func TestBlessRegen(t *testing.T) {
	// Out of combat bless is a fifth on top: mana 0.3*29 = 8.7 -> 9 plain,
	// ceil(10.44) = 11 blessed; health con 20 = 2 plain, ceil(2.4) = 3 blessed.
	if got := ManaRegen(20, 10, 9, 1, false, true); got != 11 {
		t.Errorf("blessed mana = %d, want 11", got)
	}
	if got := HealthRegen(20, 1, false, true); got != 3 {
		t.Errorf("blessed health = %d, want 3", got)
	}
	// In combat the blessed keep .65 instead of .35:
	// mana 8.7*1.2*.65 = 6.79 -> 7 against ceil(3.05) = 4 unblessed;
	// health con 40: 4*1.2*.65 = 3.12 -> 4 against ceil(1.4) = 2 unblessed.
	if got := ManaRegen(20, 10, 9, 1, true, false); got != 4 {
		t.Errorf("combat mana = %d, want 4", got)
	}
	if got := ManaRegen(20, 10, 9, 1, true, true); got != 7 {
		t.Errorf("blessed combat mana = %d, want 7", got)
	}
	if got := HealthRegen(40, 1, true, false); got != 2 {
		t.Errorf("combat health = %d, want 2", got)
	}
	if got := HealthRegen(40, 1, true, true); got != 4 {
		t.Errorf("blessed combat health = %d, want 4", got)
	}
	// Bless stacks with heal_fast: 8.7*2*1.2*.65 = 13.57 -> 14.
	if got := ManaRegen(20, 10, 9, HealFastRoomRegenMod, true, true); got != 14 {
		t.Errorf("blessed heal_fast combat mana = %d, want 14", got)
	}
}
