package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func TestEnchantedArmorMayExceedCap(t *testing.T) {
	const id = 987654
	Items[id] = &Item{Object: Object{Name: "vest"}, ItemId: id, ItemType: 5, Armor: 90, ArmorClass: 1, Flags: map[string]bool{}}
	defer delete(Items, id)
	fighter := &Character{Class: config.FIGHTER, Tier: 10}

	// Template at the t10-14 light cap (90), copy enchanted to 95.
	enchanted := &Item{Object: Object{Name: "vest"}, ItemId: id, ItemType: 5, Armor: 95, ArmorClass: 1, Flags: map[string]bool{}}
	if ok, msg := fighter.CanEquip(enchanted); !ok {
		t.Errorf("enchant bonus should be allowed past the cap: %s", msg)
	}

	// A template that is itself over the cap is still refused.
	Items[id].Armor = 95
	if ok, _ := fighter.CanEquip(enchanted); ok {
		t.Error("a piece whose own armor is over the cap should be refused")
	}
}

func TestVitalStrikeRespectsInertialBarrier(t *testing.T) {
	c := &Character{Class: config.MAGE, Tier: 20, Flags: map[string]bool{"inertial-barrier": true}, Modifiers: map[string]int{}}
	c.Int.Current = 30
	c.Vit.Current, c.Vit.Max = 1000, 1000
	c.Equipment = &Equipment{}
	// No armor, so only the barrier applies: 1000 - ceil(1000 * .1595)
	if got, _ := c.ReceiveVitalDamage(1000); got != 840 {
		t.Errorf("vital damage through barrier = %d, want 840", got)
	}
}

func TestCheckEquipmentReturnsDisallowedGear(t *testing.T) {
	helm := &Item{Object: Object{Name: "helm"}, ItemType: 25, Flags: map[string]bool{}}
	robe := &Item{Object: Object{Name: "robe"}, ItemType: 5, Flags: map[string]bool{}}
	e := &Equipment{Head: helm, Chest: robe}
	e.CanEquip = func(i *Item) (bool, string) { return i.ItemType != 25, "" }
	var returned []*Item
	e.ReturnToInventory = func(i *Item) { returned = append(returned, i) }
	e.CheckEquipment()
	if e.Head != (*Item)(nil) || e.Chest != robe {
		t.Error("helm should be removed and robe kept")
	}
	if len(returned) != 1 || returned[0] != helm {
		t.Errorf("helm not returned to inventory: %v", returned)
	}
}
