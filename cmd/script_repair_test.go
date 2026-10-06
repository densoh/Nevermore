package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
)

func TestRepairQuote(t *testing.T) {
	const id = 999901
	objects.Items[id] = &objects.Item{ItemId: id, ItemType: 0, MaxUses: 100, Value: 1000, Flags: map[string]bool{}}
	defer delete(objects.Items, id)

	weapon := func(uses, repairs int, flags map[string]bool) *objects.Item {
		if flags == nil {
			flags = map[string]bool{}
		}
		return &objects.Item{ItemId: id, ItemType: 0, MaxUses: uses, Value: 1000, Repairs: repairs, Flags: flags}
	}

	cases := []struct {
		name     string
		item     *objects.Item
		cost     int
		overhaul bool
		problem  bool
	}{
		{"half worn", weapon(50, 0, nil), 150, false, false},
		{"last repair before overhaul", weapon(50, config.RepairsBeforeOverhaul-1, nil), 150, false, false},
		{"overhaul due", weapon(50, config.RepairsBeforeOverhaul, nil), 2250, true, false},
		{"quest loot never overhauls", weapon(50, config.RepairsBeforeOverhaul+4, map[string]bool{"quest_loot": true}), 225, false, false},
		{"pristine", weapon(100, 0, nil), 0, false, true},
		{"always crit", weapon(50, 0, map[string]bool{"always_crit": true}), 0, false, true},
		{"armor overhaul due", &objects.Item{ItemId: id, ItemType: 5, MaxUses: 50, Value: 1000, Repairs: config.RepairsBeforeOverhaul, Flags: map[string]bool{}}, 2250, true, false},
	}
	for _, c := range cases {
		cost, overhaul, problem := repairQuote(c.item)
		if (problem != "") != c.problem {
			t.Errorf("%s: problem = %q", c.name, problem)
			continue
		}
		if !c.problem && (cost != c.cost || overhaul != c.overhaul) {
			t.Errorf("%s: got cost %d overhaul %v, want %d %v", c.name, cost, overhaul, c.cost, c.overhaul)
		}
	}
}
