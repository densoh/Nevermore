package config

import "testing"

func TestOutlevelDamageMult(t *testing.T) {
	cases := []struct {
		diff, engaged int
		want          float64
	}{
		{-3, 1, 1},
		{2, 1, 1},
		{3, 1, 1.05},
		{4, 1, 1.10},
		{6, 1, 1.20},
		{3, 2, 1},
		{4, 2, 1.05},
		{5, 3, 1.05},
	}
	for _, c := range cases {
		if got := OutlevelDamageMult(c.diff, c.engaged); got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("OutlevelDamageMult(%d, %d) = %v, want %v", c.diff, c.engaged, got, c.want)
		}
	}
}

func TestOutlevelMissPenalty(t *testing.T) {
	cases := []struct {
		diff, engaged, want int
	}{
		{-3, 1, 0},
		{1, 1, 0},
		{2, 1, MissPerLevel},
		{4, 1, 3 * MissPerLevel},
		{2, 2, 0},
		{3, 2, MissPerLevel},
		{4, 3, MissPerLevel},
	}
	for _, c := range cases {
		if got := OutlevelMissPenalty(c.diff, c.engaged); got != c.want {
			t.Errorf("OutlevelMissPenalty(%d, %d) = %d, want %d", c.diff, c.engaged, got, c.want)
		}
	}
}
