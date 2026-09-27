package objects

import "testing"

func greatsword() *Item {
	return &Item{Object: Object{Name: "greatsword"}, ItemType: 0, Flags: map[string]bool{"two_handed": true}}
}

func shield() *Item {
	return &Item{Object: Object{Name: "shield"}, ItemType: 23, Flags: map[string]bool{}}
}

func TestTwoHandedNeedsOffHandEmpty(t *testing.T) {
	e := &Equipment{}
	if !e.Equip(shield(), 0) {
		t.Fatal("shield should equip into an empty off hand")
	}
	if ok, msg := e.HandsFree(greatsword()); ok || msg == "" {
		t.Error("two-handed weapon allowed with a shield in the off hand")
	}
	if e.Equip(greatsword(), 0) {
		t.Error("Equip bypassed the hand check")
	}
	e.UnequipSpecific("off")
	if !e.Equip(greatsword(), 0) {
		t.Error("two-handed weapon refused with both hands free")
	}
}

func TestTwoHandedBlocksOffHand(t *testing.T) {
	e := &Equipment{}
	if !e.Equip(greatsword(), 0) {
		t.Fatal("greatsword should equip")
	}
	for _, item := range []*Item{shield(), {Object: Object{Name: "rope"}, ItemType: 13, Flags: map[string]bool{}}} {
		if ok, _ := e.HandsFree(item); ok {
			t.Errorf("%s allowed in the off hand while wielding two-handed", item.Name)
		}
		if e.Equip(item, 0) {
			t.Errorf("%s equipped in the off hand while wielding two-handed", item.Name)
		}
	}
	// Anything not held in a hand is unaffected.
	if !e.Equip(&Item{Object: Object{Name: "helm"}, ItemType: 25, Flags: map[string]bool{}}, 0) {
		t.Error("head slot blocked by a two-handed weapon")
	}
}

func TestOneHandedStillPairsWithOffHand(t *testing.T) {
	e := &Equipment{}
	sword := &Item{Object: Object{Name: "sword"}, ItemType: 0, Flags: map[string]bool{}}
	if !e.Equip(sword, 0) || !e.Equip(shield(), 0) {
		t.Error("one-handed weapon and shield should pair")
	}
}

func TestCheckEquipmentDropsOffHandForTwoHander(t *testing.T) {
	e := &Equipment{Main: greatsword(), Off: shield()}
	e.CanEquip = func(*Item) (bool, string) { return true, "" }
	var returned []*Item
	e.ReturnToInventory = func(i *Item) { returned = append(returned, i) }
	e.CheckEquipment()
	if e.Off != (*Item)(nil) || e.Main == (*Item)(nil) {
		t.Error("off hand should be cleared and the weapon kept")
	}
	if len(returned) != 1 || returned[0].Name != "shield" {
		t.Errorf("shield not returned to inventory: %v", returned)
	}
}
