package objects

import "testing"

func TestNameMatchQuality(t *testing.T) {
	cases := []struct {
		name, alias string
		want        int
	}{
		{"male wolverine", "male", 2},
		{"female wolverine", "male", 1},
		{"Female Wolverine", "FEM", 2},
		{"female wolverine", "male wol", 1},
		{"male wolverine", "male wol", 2},
		{"heal-all potion", "all", 2},
		{"Werewolf", "wolf", 1},
		{"Werewolf", "bear", 0},
	}
	for _, c := range cases {
		if got := nameMatchQuality(c.name, c.alias); got != c.want {
			t.Errorf("nameMatchQuality(%q, %q) = %d; want %d", c.name, c.alias, got, c.want)
		}
	}
}

func TestMobSearchPrefersWordStart(t *testing.T) {
	mk := func(n string) *Mob {
		m := &Mob{Flags: map[string]bool{}}
		m.Name = n
		return m
	}
	female, male, were := mk("female wolverine"), mk("male wolverine"), mk("Werewolf")
	inv := &MobInventory{Contents: []*Mob{female, male, were}}
	obs := &Character{Flags: map[string]bool{}}

	if got := inv.Search("male", 1, obs); got != male {
		t.Errorf("male -> %v", got)
	}
	if got := inv.Search("male", 2, obs); got != nil {
		t.Errorf("male 2 -> %v; want nil", got)
	}
	if got := inv.Search("wolv", 2, obs); got != male {
		t.Errorf("wolv 2 -> %v", got)
	}
	if got := inv.Search("wolf", 1, obs); got != were {
		t.Errorf("wolf -> %v", got)
	}
}

func TestItemSearchPrefersWordStart(t *testing.T) {
	mk := func(n string) *Item {
		i := &Item{Flags: map[string]bool{}}
		i.Name = n
		return i
	}
	female, male, long := mk("female idol"), mk("male idol"), mk("longsword")
	inv := &ItemInventory{Contents: []*Item{female, male, long}}

	if got := inv.Search("male", 1); got != male {
		t.Errorf("male -> %v", got)
	}
	if got := inv.Search("sword", 1); got != long {
		t.Errorf("sword -> %v", got)
	}

	eq := &Equipment{Neck: female, Ring1: male}
	if got := eq.Search("male", 1); got != male {
		t.Errorf("equipment male -> %v", got)
	}
	if slot := eq.FindLocation("male"); slot != "ring1" {
		t.Errorf("FindLocation(male) = %q; want ring1", slot)
	}
	if slot := eq.FindLocation("idol"); slot != "ring1" {
		t.Errorf("FindLocation(idol) = %q; want ring1 (last match, as before)", slot)
	}
}

func TestCharSearchPrefersWordStart(t *testing.T) {
	mk := func(n string) *Character {
		c := &Character{Flags: map[string]bool{}}
		c.Name = n
		return c
	}
	emma, ma := mk("Emma"), mk("Mal")
	inv := &CharInventory{Contents: []*Character{emma, ma}}

	if got := inv.SearchAll("ma"); got != ma {
		t.Errorf("SearchAll(ma) -> %v", got)
	}
	if got := inv.Search("ma", mk("observer")); got != ma {
		t.Errorf("Search(ma) -> %v", got)
	}
	if got := inv.Search("mm", mk("observer")); got != emma {
		t.Errorf("Search(mm) -> %v", got)
	}
}

func TestFindExitPrefersShortestWordStart(t *testing.T) {
	const from, to = 990001, 990002
	Rooms[to] = &Room{RoomId: to, Flags: map[string]bool{"active": true}}
	defer delete(Rooms, to)

	mk := func(n string) *Exit {
		e := &Exit{ParentId: from, ToId: to, Flags: map[string]bool{}}
		e.Name = n
		return e
	}
	west, northwest := mk("west"), mk("northwest")
	r := &Room{RoomId: from, Exits: map[string]*Exit{"west": west, "northwest": northwest}}
	obs := &Character{Flags: map[string]bool{}}

	// map order is random, so try enough times to catch an order-dependent pick
	for range 50 {
		if got := r.FindExit("west", obs); got != west {
			t.Fatalf("west -> %v", got.Name)
		}
		if got := r.FindExit("w", obs); got != west {
			t.Fatalf("w -> %v", got.Name)
		}
		if got := r.FindExit("north", obs); got != northwest {
			t.Fatalf("north -> %v", got.Name)
		}
	}
}
