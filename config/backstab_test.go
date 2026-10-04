package config

import (
	"math"
	"testing"
)

func TestBackstabMultiplier(t *testing.T) {
	// 2.5x at stealth 0, 0.2 more per level, plus 1 at grandmaster: 5.5x.
	cases := map[int]float64{0: 2.5, 7: 3.9, 9: 4.3, 10: 5.5}
	for level, want := range cases {
		if got := BackstabMultiplier(level); math.Abs(got-want) > 1e-9 {
			t.Errorf("stealth %d multiplier = %v, want %v", level, got, want)
		}
	}
}

func TestBackstabMissPenaltyFor(t *testing.T) {
	// 30 points over a regular swing at stealth 0, 2 fewer per level, 10 at grandmaster.
	cases := map[int]int{0: 30, 5: 20, 10: 10}
	for level, want := range cases {
		if got := BackstabMissPenaltyFor(level); got != want {
			t.Errorf("stealth %d penalty = %d, want %d", level, got, want)
		}
	}
}
