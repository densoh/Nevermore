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
