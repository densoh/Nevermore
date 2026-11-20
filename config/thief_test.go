package config

import "testing"

// The dice match the agreed table: fractions sum before truncating.
func TestSneakAttackDice(t *testing.T) {
	cases := []struct{ tier, rank, want int }{
		{9, 10, 0},   // below the tier gate
		{10, 0, 1},   // 1
		{10, 2, 2},   // 1 + 1
		{12, 1, 2},   // 0.67 + 1 + 0.5 = 2.17
		{13, 3, 3},   // 1 + 1 + 1.5 = 3.5
		{16, 5, 5},   // 2 + 1 + 2.5 = 5.5
		{20, 7, 7},   // 3.33 + 1 + 3.5 = 7.83
		{25, 10, 11}, // 5 + 1 + 5
	}
	for _, c := range cases {
		if got := SneakAttackDice(c.tier, c.rank); got != c.want {
			t.Errorf("tier %d rank %d: %d dice, want %d", c.tier, c.rank, got, c.want)
		}
	}
}

func TestSneakAttackChance(t *testing.T) {
	for attackers, want := range map[int]int{0: 100, 1: 75, 2: 50, 3: 25, 4: 0, 6: 0} {
		if got := SneakAttackChance(attackers); got != want {
			t.Errorf("%d other attackers: %d%%, want %d%%", attackers, got, want)
		}
	}
}

// Poison dice match the agreed table.
func TestPoisonStrikeDice(t *testing.T) {
	cases := []struct{ tier, rank, want int }{
		{10, 0, 1},
		{10, 2, 1},  // 1 + 0.5
		{13, 3, 2},  // 1 + 0.75 + 0.6 = 2.35
		{16, 5, 3},  // 1 + 1.25 + 1.2 = 3.45
		{20, 7, 4},  // 1 + 1.75 + 2 = 4.75
		{25, 10, 6}, // 1 + 2.5 + 3 = 6.5
	}
	for _, c := range cases {
		if got := PoisonStrikeDice(c.tier, c.rank); got != c.want {
			t.Errorf("tier %d rank %d: %d dice, want %d", c.tier, c.rank, got, c.want)
		}
	}
}

func TestDeviousFromSpell(t *testing.T) {
	cases := map[string]string{
		"devious-apply-poison":    DeviousPoison,
		"devious-expose-weakness": DeviousExpose,
	}
	for spell, want := range cases {
		if got, ok := DeviousFromSpell(spell); !ok || got != want {
			t.Errorf("%q: %q %v, want %q", spell, got, ok, want)
		}
		if DeviousSpell(want) != spell {
			t.Errorf("DeviousSpell(%q) = %q, want %q", want, DeviousSpell(want), spell)
		}
	}
	for _, spell := range []string{"poison", "devious-poison", "devious-silence", "fireball", ""} {
		if _, ok := DeviousFromSpell(spell); ok {
			t.Errorf("%q read as a devious strike", spell)
		}
	}
}
