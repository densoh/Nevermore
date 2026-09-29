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
	// rolled when the human floor was 5 across the board
	short := StatShortfalls(HUMAN, 20, 10, 20, 5, 5)
	if len(short) != 1 || short["pie"] != 1 {
		t.Errorf("fighter with pie 5 shortfalls = %v, want pie 1", short)
	}
	short = StatShortfalls(HUMAN, 5, 5, 5, 15, 20)
	if len(short) != 1 || short["con"] != 1 {
		t.Errorf("mage with con 5 shortfalls = %v, want con 1", short)
	}
	if short = StatShortfalls(HUMAN, 5, 5, 6, 28, 6); len(short) != 0 {
		t.Errorf("stats on the minimum shortfalls = %v, want none", short)
	}
	// the same stats are fine on a race with lower floors
	if short = StatShortfalls(HALF_ORC, 20, 10, 20, 5, 5); len(short) != 0 {
		t.Errorf("half-orc shortfalls = %v, want none", short)
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
