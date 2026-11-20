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

func TestLevelCapIgnoresGearCaps(t *testing.T) {
	const id = 987655
	Items[id] = &Item{Object: Object{Name: "plate"}, ItemId: id, ItemType: 5, Armor: 500, ArmorClass: 3, Flags: map[string]bool{}}
	defer delete(Items, id)
	plate := &Item{Object: Object{Name: "plate"}, ItemId: id, ItemType: 5, Armor: 500, ArmorClass: 3, Flags: map[string]bool{}}
	sword := &Item{Object: Object{Name: "sword"}, ItemType: 0, SidesDice: 100, NumDice: 2, Flags: map[string]bool{}}

	below := &Character{Class: config.FIGHTER, Tier: config.LevelCap - 1}
	if ok, _ := below.CanEquip(plate); ok {
		t.Error("over-cap armor should be refused below LevelCap")
	}
	if ok, _ := below.CanEquip(sword); ok {
		t.Error("over-cap weapon should be refused below LevelCap")
	}

	capped := &Character{Class: config.FIGHTER, Tier: config.LevelCap}
	if ok, msg := capped.CanEquip(plate); !ok {
		t.Errorf("over-cap armor should be allowed at LevelCap: %s", msg)
	}
	if ok, msg := capped.CanEquip(sword); !ok {
		t.Errorf("over-cap weapon should be allowed at LevelCap: %s", msg)
	}

	// Class rules still hold at LevelCap.
	mage := &Character{Class: config.MAGE, Tier: config.LevelCap}
	if ok, _ := mage.CanEquip(plate); ok {
		t.Error("armor class restriction should still apply at LevelCap")
	}
}

func TestVitalStrikeRespectsInertialBarrier(t *testing.T) {
	c := &Character{Class: config.MAGE, Tier: 20, Flags: map[string]bool{"inertial-barrier": true}, Modifiers: map[string]int{}}
	c.Int.Current = 30
	c.Vit.Current, c.Vit.Max = 1000, 1000
	c.Equipment = &Equipment{}
	// No armor, so only the barrier applies: 1000 - ceil(1000 * .1595)
	// and the barrier's 160 shows up as resisted.
	if got, resist := c.ReceiveVitalDamage(1000); got != 840 || resist != 160 {
		t.Errorf("vital damage through barrier = %d taken, %d resisted; want 840, 160", got, resist)
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

// A mob's barrier cuts the hit before armor, so the two multiply: 100 less
// 15% is 85, and 50 armor resists a quarter of that (21), leaving 64. The
// resisted total counts the barrier's 15 too.
func TestMobInertialBarrierMultipliesWithArmor(t *testing.T) {
	m := &Mob{Armor: 50, Stam: Meter{Max: 1000, Current: 1000}, Flags: map[string]bool{"inertial-barrier": true}}
	got, _, resist := m.ReceiveDamage(100)
	if got != 64 || resist != 36 {
		t.Errorf("mob barrier + 50 armor on 100 = %d taken, %d resisted; want 64, 36", got, resist)
	}
}

// A naked tier 12 barbarian (con 25, dex 18) gets 41.5% of the seven slot
// caps (340, rings and shield excluded) = 141.1, times the 1.1875 con bonus.
func TestBarbarianUnarmoredArmor(t *testing.T) {
	c := &Character{Class: config.BARBARIAN, Tier: 12, Modifiers: map[string]int{}, Equipment: &Equipment{}}
	c.Con.Current, c.Dex.Current = 25, 18
	if got := c.GetStat("armor"); got != 167 {
		t.Errorf("naked barbarian armor = %d, want 167", got)
	}
	if got := config.UnarmoredMiss(c.Equipment.EmptyUnarmoredSlots()); got != 8 {
		t.Errorf("naked barbarian mob to-hit cut = %d, want 8", got)
	}
	// Wearing body armor removes that slot's share and its 2 points.
	c.Equipment.Chest = &Item{ItemType: 5, Armor: 76}
	c.Equipment.Armor = 76
	if got := config.UnarmoredMiss(c.Equipment.EmptyUnarmoredSlots()); got != 6 {
		t.Errorf("barbarian in body armor to-hit cut = %d, want 6", got)
	}
	// (76 + 41.5% of 250) * 1.1875 = 213.45
	if got := c.GetStat("armor"); got != 213 {
		t.Errorf("barbarian in body armor = %d, want 213", got)
	}
	// A worn piece with no armor still counts as empty.
	c.Equipment.Head = &Item{ItemType: 25, Armor: 0}
	if got := config.UnarmoredMiss(c.Equipment.EmptyUnarmoredSlots()); got != 6 {
		t.Errorf("barbarian in body armor and a 0-armor hat to-hit cut = %d, want 6", got)
	}
	// Other classes get nothing for empty slots.
	f := &Character{Class: config.FIGHTER, Tier: 12, Modifiers: map[string]int{}, Equipment: &Equipment{}}
	f.Con.Current, f.Dex.Current = 25, 18
	if got := f.GetStat("armor"); got != 0 {
		t.Errorf("naked fighter armor = %d, want 0", got)
	}
}
