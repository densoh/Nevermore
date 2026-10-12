package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func TestMageArmorSlots(t *testing.T) {
	c := &Character{Class: config.MAGE, Tier: 10}
	allowed := map[int]bool{5: true, 22: true, 24: true, 6: true, 8: true, 12: true, 13: true}
	for _, itemType := range []int{5, 6, 8, 12, 13, 19, 20, 21, 22, 23, 24, 25, 26} {
		item := &Item{Object: Object{Name: "thing"}, ItemType: itemType, Flags: map[string]bool{}}
		ok, _ := c.CanEquip(item)
		if ok != allowed[itemType] {
			t.Errorf("mage CanEquip type %d (%s) = %v, want %v", itemType, config.ItemTypes[itemType], ok, allowed[itemType])
		}
	}
}

func TestArmorClassRestrictions(t *testing.T) {
	const light, heavy, robe = 1, 3, 4
	want := map[int]map[int]bool{
		config.FIGHTER:   {light: true, heavy: true, robe: true},
		config.PALADIN:   {light: true, heavy: true, robe: true},
		config.BARBARIAN: {light: true, robe: true},
		config.CLERIC:    {light: true, robe: true},
		config.THIEF:     {light: true, robe: true},
		config.RANGER:    {light: true, robe: true},
		config.BARD:      {light: true, robe: true},
		config.MAGE:      {robe: true},
		config.MONK:      {robe: true},
	}
	for class, allowed := range want {
		c := &Character{Class: class, Tier: 10}
		for _, armorClass := range []int{light, heavy, robe} {
			item := &Item{Object: Object{Name: "armor"}, ItemType: 5, ArmorClass: armorClass, Flags: map[string]bool{}}
			ok, _ := c.CanEquip(item)
			if ok != allowed[armorClass] {
				t.Errorf("%s CanEquip %s body = %v, want %v", config.AvailableClasses[class], config.ArmorClass[armorClass], ok, allowed[armorClass])
			}
		}
		cloak := &Item{Object: Object{Name: "cloak"}, ItemType: 22, Flags: map[string]bool{}}
		if ok, _ := c.CanEquip(cloak); !ok {
			t.Errorf("%s should wear an unclassed cloak", config.AvailableClasses[class])
		}
	}
}

func TestArmorStatUsesLightPieces(t *testing.T) {
	chest := &Item{Object: Object{Name: "plate"}, ItemType: 5, Armor: 100, ArmorClass: 3, Flags: map[string]bool{}}
	legs := &Item{Object: Object{Name: "leathers"}, ItemType: 20, Armor: 60, ArmorClass: 1, Flags: map[string]bool{}}
	ring := &Item{Object: Object{Name: "ring"}, ItemType: 24, Armor: 10, Flags: map[string]bool{}}
	shield := &Item{Object: Object{Name: "shield"}, ItemType: 23, Armor: 80, Flags: map[string]bool{}}
	c := &Character{Class: config.FIGHTER, Tier: 15, Modifiers: map[string]int{}}
	c.Con.Current, c.Dex.Current = 20, 20
	// Heavy chest and the shield are out; light legs and the unclassed ring are in.
	c.Equipment = &Equipment{Armor: 250, Chest: chest, Legs: legs, Ring1: ring, Off: shield}
	if got := c.Equipment.LightArmor(); got != 70 {
		t.Fatalf("LightArmor = %d, want 70", got)
	}
	// 250 + 250*20*.005 (25) + 70*20*.005 (7)
	if got := c.GetStat("armor"); got != 282 {
		t.Errorf("fighter armor = %d, want 282", got)
	}

	// A robe counts as lighter than light.
	robe := &Item{Object: Object{Name: "robe"}, ItemType: 5, Armor: 100, ArmorClass: 4, Flags: map[string]bool{}}
	c.Equipment = &Equipment{Armor: 170, Chest: robe, Legs: legs, Ring1: ring}
	if got := c.Equipment.LightArmor(); got != 170 {
		t.Errorf("LightArmor with a robe = %d, want 170", got)
	}
}

func TestHeavyArmorNeedsTierTen(t *testing.T) {
	plate := &Item{Object: Object{Name: "plate"}, ItemType: 5, Armor: 100, ArmorClass: 3, Flags: map[string]bool{}}
	if ok, _ := (&Character{Class: config.FIGHTER, Tier: 9}).CanEquip(plate); ok {
		t.Error("a tier 9 fighter should not wear heavy armor")
	}
	// 100 is over the t10-14 light cap (90) but within the heavy one (110).
	if ok, msg := (&Character{Class: config.FIGHTER, Tier: 10}).CanEquip(plate); !ok {
		t.Errorf("a tier 10 fighter should wear 100 armor plate: %s", msg)
	}
}

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
	// No armor, so only the barrier applies: 1000 - ceil(1000 * .2375)
	if got, _ := c.ReceiveVitalDamage(1000); got != 762 {
		t.Errorf("vital damage through barrier = %d, want 762", got)
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
