package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/objects"
)

func TestMeldQuote(t *testing.T) {
	const ammoId, deviceId, bigDeviceId = 999902, 999903, 999904
	objects.Items[ammoId] = &objects.Item{ItemId: ammoId, ItemType: 15, MaxUses: 20, Value: 600}
	objects.Items[deviceId] = &objects.Item{ItemId: deviceId, ItemType: 6, MaxUses: 10}
	objects.Items[bigDeviceId] = &objects.Item{ItemId: bigDeviceId, ItemType: 6, MaxUses: 30}
	defer delete(objects.Items, ammoId)
	defer delete(objects.Items, deviceId)
	defer delete(objects.Items, bigDeviceId)

	// device has 10 base uses; uses above that were melded on earlier.
	device := func(spell string, uses int) *objects.Item {
		return &objects.Item{ItemId: deviceId, ItemType: 6, Spell: spell, MaxUses: uses}
	}
	ammo := func(uses int) *objects.Item {
		return &objects.Item{ItemId: ammoId, ItemType: 15, MaxUses: uses}
	}

	cases := []struct {
		name         string
		target, meld *objects.Item
		cost         int
		problem      bool
	}{
		{"vigor, tier 1", device("vigor", 10), device("vigor", 10), 40, false},
		{"detraumatize, tier 7", device("detraumatize", 10), device("detraumatize", 10), 520, false},
		{"renewal, tier 11", device("renewal", 10), device("renewal", 10), 1240, false},
		{"exactly 100 uses, no penalty", device("renewal", 90), device("renewal", 10), 1240, false},
		{"101 uses, +50%", device("renewal", 91), device("renewal", 10), 1860, false},
		{"200 uses, +100%", device("renewal", 190), device("renewal", 10), 2480, false},
		{"300 uses, +150%", device("renewal", 290), device("renewal", 10), 3100, false},
		{"melded item into fresh one pays base uses only", device("renewal", 10), device("renewal", 50), 1240, false},
		{"two melded items pay base uses only", device("renewal", 50), device("renewal", 50), 1240, false},
		{"larger base uses wins, either order", device("renewal", 10), &objects.Item{ItemId: bigDeviceId, ItemType: 6, Spell: "renewal", MaxUses: 30}, 3720, false},
		{"larger base uses wins, reversed", &objects.Item{ItemId: bigDeviceId, ItemType: 6, Spell: "renewal", MaxUses: 30}, device("renewal", 10), 3720, false},
		{"ammo uses template value", ammo(20), ammo(20), 200, false},
		{"different spells", device("vigor", 10), device("renewal", 10), 0, true},
		{"different types", device("vigor", 10), &objects.Item{ItemType: 8, Spell: "vigor", MaxUses: 10}, 0, true},
		{"not meldable", &objects.Item{ItemType: 0, MaxUses: 10}, &objects.Item{ItemType: 0, MaxUses: 10}, 0, true},
	}
	for _, c := range cases {
		cost, problem := meldQuote(c.target, c.meld)
		if (problem != "") != c.problem {
			t.Errorf("%s: problem = %q", c.name, problem)
			continue
		}
		if !c.problem && cost != c.cost {
			t.Errorf("%s: got cost %d, want %d", c.name, cost, c.cost)
		}
	}

	same := device("vigor", 10)
	if _, problem := meldQuote(same, same); problem == "" {
		t.Error("melding an item into itself should be refused")
	}
}
