package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func TestMageArmorSlots(t *testing.T) {
	c := &Character{Class: config.MAGE, Tier: 10}
	const robe = 4
	cases := []struct {
		itemType, armorClass int
		want                 bool
	}{
		{5, robe, true}, // robe
		{5, 0, false},   // unclassed body armor
		{5, 1, false},   // light body armor
		{22, 0, true},   // cloak
		{24, 0, true},   // ring
		{23, 0, false},  // shield
		{19, 0, false},  // head
		{20, 0, false},  // legs
		{21, 0, false},  // feet
		{25, 0, false},  // arms
		{26, 0, false},  // hands
		{6, 0, true},    // off hand item that isn't armor
	}
	for _, tc := range cases {
		item := &Item{Object: Object{Name: "thing"}, ItemType: tc.itemType, ArmorClass: tc.armorClass, Flags: map[string]bool{}}
		ok, _ := c.CanEquip(item)
		if ok != tc.want {
			t.Errorf("mage CanEquip type %d (%s) class %d = %v, want %v", tc.itemType, config.ItemTypes[tc.itemType], tc.armorClass, ok, tc.want)
		}
	}
}

func TestEveryClassCanWearRobes(t *testing.T) {
	for class := config.FIGHTER; class <= config.MONK; class++ {
		if class == config.MONK {
			continue // monks can't wear armor with any value
		}
		c := &Character{Class: class, Tier: 10}
		item := &Item{Object: Object{Name: "robe"}, ItemType: 5, ArmorClass: 4, Flags: map[string]bool{}}
		if ok, msg := c.CanEquip(item); !ok {
			t.Errorf("class %d can't wear a robe: %s", class, msg)
		}
	}
}
