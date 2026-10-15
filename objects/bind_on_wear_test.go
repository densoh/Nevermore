package objects

import "testing"

func TestDamageBindsRingToWearer(t *testing.T) {
	ring := &Item{Object: Object{Name: "ring"}, ItemType: 24, Armor: 5, MaxUses: 50, Flags: map[string]bool{}}
	e := &Equipment{Owner: "Alice", Ring1: ring}
	if msg := e.DamageRandomArmor(); msg == "" {
		t.Error("first damage should announce the bind")
	}
	if ring.BoundTo != "Alice" || ring.Flags["permanent"] || ring.MaxUses != 49 {
		t.Errorf("ring bound to %q, permanent %v, uses %d", ring.BoundTo, ring.Flags["permanent"], ring.MaxUses)
	}
	if msg := e.DamageRandomArmor(); msg != "" {
		t.Errorf("already bound ring should be quiet, got %q", msg)
	}
}

func TestDamageSoulbindsQuestLoot(t *testing.T) {
	vest := &Item{Object: Object{Name: "vest"}, ItemType: 5, Armor: 50, MaxUses: 50, Flags: map[string]bool{"quest_loot": true}}
	e := &Equipment{Owner: "Alice", Chest: vest}
	e.DamageRandomArmor()
	if !vest.Flags["permanent"] || vest.BoundTo != "" {
		t.Errorf("quest loot permanent %v, bound to %q", vest.Flags["permanent"], vest.BoundTo)
	}

	sword := &Item{Object: Object{Name: "sword"}, MaxUses: 50, Flags: map[string]bool{"quest_loot": true}}
	e = &Equipment{Owner: "Alice", Main: sword}
	if e.DamageWeapon("main", 0); sword.Flags["permanent"] {
		t.Error("a hit that does no weapon damage shouldn't bind")
	}
	if msg := e.DamageWeapon("main", 1); msg == "" || !sword.Flags["permanent"] {
		t.Errorf("quest loot weapon should soulbind on damage, msg %q", msg)
	}
}

func TestDamageLeavesOrdinaryArmorUnbound(t *testing.T) {
	vest := &Item{Object: Object{Name: "vest"}, ItemType: 5, Armor: 50, MaxUses: 50, Flags: map[string]bool{}}
	e := &Equipment{Owner: "Alice", Chest: vest}
	if msg := e.DamageRandomArmor(); msg != "" || vest.Bound() {
		t.Errorf("ordinary armor bound: msg %q", msg)
	}
}

func TestBindingSurvivesSaveAndRestore(t *testing.T) {
	const id = 999902
	Items[id] = &Item{Object: Object{Name: "ring"}, ItemId: id, ItemType: 24, Armor: 5, MaxUses: 50, Flags: map[string]bool{}}
	defer delete(Items, id)

	ring := &Item{Object: Object{Name: "ring"}, ItemId: id, ItemType: 24, Armor: 5, MaxUses: 40, Flags: map[string]bool{}, BoundTo: "Alice"}
	soul := &Item{Object: Object{Name: "ring"}, ItemId: id, ItemType: 24, Armor: 5, MaxUses: 40, Flags: map[string]bool{"permanent": true}}
	inv := &ItemInventory{}
	inv.Add(ring)
	inv.Add(soul)

	restored := RestoreInventory(inv.Jsonify()).Contents
	if len(restored) != 2 {
		t.Fatalf("restored %d items", len(restored))
	}
	bound, perm := 0, 0
	for _, item := range restored {
		if item.BoundTo == "Alice" {
			bound++
		}
		if item.Flags["permanent"] {
			perm++
		}
	}
	if bound != 1 || perm != 1 {
		t.Errorf("after restore: %d bound to Alice, %d permanent", bound, perm)
	}

	equip, errored := RestoreEquipment(`[{"itemId":999902,"name":"ring","uses":40,"magic":0,"spell":"","armor":5,"bound_to":"Alice"}]`, 0)
	if len(errored) != 0 || equip.Ring1 == (*Item)(nil) || equip.Ring1.BoundTo != "Alice" {
		t.Errorf("equipped ring lost its binding: %+v", equip.Ring1)
	}
}
