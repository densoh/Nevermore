package config

import (
	"math"
	"testing"
)

func TestHamstringDamage(t *testing.T) {
	cases := map[int]int{1: 1, 2: 1, 3: 2, 10: 5, 11: 6, 40: 20}
	for roll, want := range cases {
		if got := HamstringDamage(roll); got != want {
			t.Errorf("HamstringDamage(%d) = %d, want %d", roll, got, want)
		}
	}
}

func TestCrippledChance(t *testing.T) {
	cases := map[int]int{85: 42, 40: 20, 35: 17, 0: 0}
	for chance, want := range cases {
		if got := CrippledChance(chance); got != want {
			t.Errorf("CrippledChance(%d) = %d, want %d", chance, got, want)
		}
	}
}

func TestCircleStunFor(t *testing.T) {
	cases := []struct{ class, level, want int }{
		{FIGHTER, 0, 1},
		{FIGHTER, 1, 1},
		{FIGHTER, 2, 2},
		{FIGHTER, 5, 3},
		{FIGHTER, 10, 6},
		{BARBARIAN, 10, 1},
	}
	for _, c := range cases {
		if got := CircleStunFor(c.class, c.level); got != c.want {
			t.Errorf("CircleStunFor(%d, %d) = %d, want %d", c.class, c.level, got, c.want)
		}
	}
}

func TestShieldSlamStun(t *testing.T) {
	if got := ShieldSlamStun(PALADIN, 40, 20); got != 8 {
		t.Errorf("paladin ShieldSlamStun = %d, want 8 (piety)", got)
	}
	if got := ShieldSlamStun(FIGHTER, 40, 20); got != 16 {
		t.Errorf("fighter ShieldSlamStun = %d, want 16 (strength)", got)
	}
}

func TestExecuteMultiplier(t *testing.T) {
	cases := []struct {
		tier, level int
		health      float64
		want        float64
	}{
		{15, 7, 1.0, 1},    // full health
		{15, 7, 0.5, 1},    // at the start line, nothing yet
		{15, 7, 0.3, 1.5},  // halfway down the window
		{15, 7, 0.2, 1.75}, // three quarters
		{15, 7, 0.1, 2},    // window fully open
		{15, 7, 0.02, 2},   // stays capped below it
		{14, 10, 0.05, 1},  // tier too low
		{20, 6, 0.05, 1},   // adept, one level short of expert
		{20, 10, 0.05, 2},  // grandmaster
	}
	for _, c := range cases {
		got := ExecuteMultiplier(c.tier, c.level, c.health)
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("ExecuteMultiplier(%d, %d, %.2f) = %.3f, want %.3f", c.tier, c.level, c.health, got, c.want)
		}
	}
}

func TestParryChance(t *testing.T) {
	cases := []struct {
		level, tier int
		shield      bool
		want        float64
	}{
		{0, 20, false, 0},
		{0, 20, true, 0}, // no skill, no parry at all
		{1, 1, false, 1.5},
		{1, 1, true, 2},
		{7, 15, false, 10.5},
		{7, 15, true, 18},
		{10, 20, false, 15},
		{10, 20, true, 25},
	}
	for _, c := range cases {
		if got := ParryChance(c.level, c.tier, c.shield); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("ParryChance(%d, %d, %v) = %.1f, want %.1f", c.level, c.tier, c.shield, got, c.want)
		}
	}
}

func TestRollFighterParryShieldBlocksNeverRiposte(t *testing.T) {
	// Over many rolls: without a shield no parry is ever a shield block, and
	// with one the block share stays inside the bonus window.
	for i := 0; i < 2000; i++ {
		if _, block := RollFighterParry(10, 20, false); block {
			t.Fatal("shield block reported with no shield")
		}
	}
	parries, blocks := 0, 0
	for i := 0; i < 20000; i++ {
		p, b := RollFighterParry(10, 20, true)
		if p {
			parries++
		}
		if b {
			blocks++
		}
	}
	if blocks == 0 || blocks >= parries {
		t.Errorf("shield blocks %d of %d parries; expected a minority but non-zero share", blocks, parries)
	}
}

func TestShieldSlamDamage(t *testing.T) {
	cases := []struct{ str, tier, roll, want int }{
		{20, 5, 0, 45},    // no shield armor
		{20, 5, 10, 50},   // half the roll
		{20, 5, 11, 50},   // odd rolls round down
		{40, 20, 60, 130}, // strong fighter, heavy shield max roll (2d30)
	}
	for _, c := range cases {
		if got := ShieldSlamDamage(c.str, c.tier, c.roll); got != c.want {
			t.Errorf("ShieldSlamDamage(%d, %d, %d) = %d, want %d", c.str, c.tier, c.roll, got, c.want)
		}
	}
	if RollShieldArmor(0) != 0 {
		t.Error("armorless shield rolled damage")
	}
	for i := 0; i < 200; i++ {
		if r := RollShieldArmor(12); r < 2 || r > 24 {
			t.Fatalf("RollShieldArmor(12) = %d, outside 2d12", r)
		}
	}
}

func TestThreatBumps(t *testing.T) {
	if got := ThreatPercent(400, TauntThreatPercent); got != 200 {
		t.Errorf("taunt threat on 400 max = %d, want 200", got)
	}
	if got := ThreatPercent(400, BashThreatPercent); got != 100 {
		t.Errorf("bash threat on 400 max = %d, want 100", got)
	}
	if got := FailedTurnThreat(350, 400); got != 200 {
		t.Errorf("failed turn at 350/400 = %d, want capped 200", got)
	}
	if got := FailedTurnThreat(120, 400); got != 120 {
		t.Errorf("failed turn at 120/400 = %d, want 120", got)
	}
}

func TestMultiAttackMissPenaltyLadder(t *testing.T) {
	want := map[int]int{5: 25, 6: 25, 7: 25, 8: 20, 9: 15, 10: 10}
	for level, penalty := range want {
		if got := MultiAttackMissPenaltyFor(level); got != penalty {
			t.Errorf("MultiAttackMissPenaltyFor(%d) = %d, want %d", level, got, penalty)
		}
	}
}
