package config

import "testing"

// A starting character has 50 points and no stat above 20, so the minimums
// have to leave room for both.
func TestRaceMinsFitCreation(t *testing.T) {
	for race, name := range AvailableRaces {
		total := 0
		for stat, min := range StatMins(race) {
			if min > 20 {
				t.Errorf("%s %s minimum %d is over the creation cap", name, stat, min)
			}
			total += min
		}
		if total > 50 {
			t.Errorf("%s minimums total %d, over the 50 points a new character has", name, total)
		}
	}
}

func TestStatShortfalls(t *testing.T) {
	// rolled when the sprite piety floor was 2
	short := StatShortfalls(SPRITE, 9, 20, 10, 7, 4)
	if len(short) != 1 || short["pie"] != 1 {
		t.Errorf("sprite with pie 4 shortfalls = %v, want pie 1", short)
	}
	short = StatShortfalls(SPRITE, 16, 20, 3, 7, 2)
	if len(short) != 2 || short["con"] != 1 || short["pie"] != 3 {
		t.Errorf("sprite with con 3 and pie 2 shortfalls = %v, want con 1 and pie 3", short)
	}
	if short = StatShortfalls(SPRITE, 1, 17, 4, 7, 5); len(short) != 0 {
		t.Errorf("stats on the minimum shortfalls = %v, want none", short)
	}
	// the same stats are fine on a race with lower floors
	if short = StatShortfalls(HUMAN, 9, 20, 10, 7, 5); len(short) != 0 {
		t.Errorf("human shortfalls = %v, want none", short)
	}
}

func TestTrainingCoversShortfalls(t *testing.T) {
	cases := []struct {
		name  string
		short map[string]int
		picks []string
		want  bool
	}{
		{"nothing short", map[string]int{}, []string{"str", "str"}, true},
		{"one short, not picked", map[string]int{"pie": 1}, []string{"str", "con"}, false},
		{"one short, picked first", map[string]int{"pie": 1}, []string{"pie", "str"}, true},
		{"one short, picked second", map[string]int{"pie": 1}, []string{"str", "pie"}, true},
		{"one short, picked twice", map[string]int{"pie": 1}, []string{"pie", "pie"}, true},
		{"two short in one stat, one pick", map[string]int{"pie": 2}, []string{"pie", "str"}, false},
		{"two short in one stat, both picks", map[string]int{"pie": 2}, []string{"pie", "pie"}, true},
		{"two stats short, one picked", map[string]int{"con": 1, "pie": 1}, []string{"pie", "str"}, false},
		{"two stats short, one picked twice", map[string]int{"con": 1, "pie": 1}, []string{"pie", "pie"}, false},
		{"two stats short, both picked", map[string]int{"con": 1, "pie": 1}, []string{"con", "pie"}, true},
		{"more short than one tier covers", map[string]int{"con": 2, "pie": 3}, []string{"pie", "pie"}, true},
		{"more short than one tier covers, one wasted", map[string]int{"con": 2, "pie": 3}, []string{"pie", "dex"}, false},
	}
	for _, c := range cases {
		if got := TrainingCoversShortfalls(c.short, c.picks); got != c.want {
			t.Errorf("%s: short %v picks %v = %v, want %v", c.name, c.short, c.picks, got, c.want)
		}
	}
}
