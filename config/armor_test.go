package config

import "testing"

func TestInertialBarrierIgnore(t *testing.T) {
	cases := []struct {
		tier, intel int
		want        float64
	}{
		{9, 20, .19},    // tier term starts above 9
		{10, 20, .1925}, // 15 + 0.25 + 4
		{15, 25, .215},  // 15 + 1.5 + 5
		{20, 30, .2375}, // 15 + 2.75 + 6
		{25, 45, .28},   // theoretical top
		{5, 20, .19},    // never negative below the start tier
	}
	for _, c := range cases {
		got := InertialBarrierIgnore(c.tier, c.intel)
		if got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("InertialBarrierIgnore(%d, %d) = %.4f, want %.4f", c.tier, c.intel, got, c.want)
		}
	}
}

func TestArmorDescription(t *testing.T) {
	cases := map[[2]int]string{
		{5, 3}:  "heavy body armor",
		{20, 1}: "light legs armor",
		{5, 4}:  "robe body armor",
		{24, 0}: "finger armor",
	}
	for in, want := range cases {
		if got := ArmorDescription(in[0], in[1]); got != want {
			t.Errorf("ArmorDescription(%d, %d) = %q, want %q", in[0], in[1], got, want)
		}
	}
}
