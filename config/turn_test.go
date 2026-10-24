package config

import "testing"

func TestTurnSuccessChance(t *testing.T) {
	cases := []struct {
		tier, mobLevel, pie, rank, want int
	}{
		{10, 10, 20, 0, 50},  // same level, 20 pie matches the old flat 50
		{10, 10, 45, 0, 62},  // max pie, odd point rounds down
		{12, 10, 30, 0, 75},  // two levels over
		{10, 12, 30, 0, 35},  // two levels under
		{10, 15, 30, 0, 5},   // five under: at the floor
		{10, 20, 30, 0, 5},   // ten under: still 5
		{10, 10, 16, 3, 51},  // tier 10 paladin, divinity rank 3
		{15, 15, 20, 7, 57},  // tier 15 healing paladin, rank 7
		{10, 10, 20, 10, 60}, // grandmaster is a full mob level
	}
	for _, c := range cases {
		if got := TurnSuccessChance(c.tier, c.mobLevel, c.pie, c.rank); got != c.want {
			t.Errorf("TurnSuccessChance(%d, %d, %d, %d) = %d, want %d", c.tier, c.mobLevel, c.pie, c.rank, got, c.want)
		}
	}
}

func TestTurnDisintegratePerMille(t *testing.T) {
	cases := []struct{ chance, want int }{
		{72, 72},   // capped cleric: 7.2%
		{50, 50},   // old flat-50 turn keeps the old 5%
		{30, 30},   // where the floor takes over
		{20, 30},   // floor: 3%
		{5, 30},    // at the success floor: 3% disintegrate, 2% half
		{130, 100}, // success clamps at 100, so 10%
	}
	for _, c := range cases {
		if got := TurnDisintegratePerMille(c.chance); got != c.want {
			t.Errorf("TurnDisintegratePerMille(%d) = %d, want %d", c.chance, got, c.want)
		}
	}
}

func TestTurnDivinityExp(t *testing.T) {
	cases := []struct {
		damage, maxStam, mobExp, want int
	}{
		{500, 1000, 4000, 1000},  // half the mob's max at half rate
		{1000, 1000, 4000, 2000}, // disintegrate on a full mob
		{250, 1000, 4000, 500},   // turn on a half-dead mob earns half
		{0, 1000, 4000, 0},
		{500, 0, 4000, 0}, // no max stamina, no exp
	}
	for _, c := range cases {
		if got := TurnDivinityExp(c.damage, c.maxStam, c.mobExp); got != c.want {
			t.Errorf("TurnDivinityExp(%d, %d, %d) = %d, want %d", c.damage, c.maxStam, c.mobExp, got, c.want)
		}
	}
}
