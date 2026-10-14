package config

import (
	"strconv"
	"strings"
)

// MaxArmor Max Damage Reduction for level group and armor type
var MaxArmor = map[int]map[int]int{
	// Level int ref is cap
	4:  {5: 50, 20: 40, 21: 15, 19: 15, 25: 30, 22: 15, 26: 15, 23: 20, 24: 4},
	9:  {5: 70, 20: 50, 21: 25, 19: 25, 25: 40, 22: 25, 26: 25, 23: 40, 24: 6},
	14: {5: 90, 20: 60, 21: 35, 19: 35, 25: 50, 22: 35, 26: 35, 23: 60, 24: 8},
	19: {5: 110, 20: 70, 21: 45, 19: 45, 25: 60, 22: 45, 26: 45, 23: 80, 24: 10},
	25: {5: 130, 20: 80, 21: 55, 19: 55, 25: 70, 22: 55, 26: 55, 23: 100, 24: 12},
	// Heavy armor only: a heavy piece reads the band above the wearer's, so
	// tier 20-25 heavy needs one more row. Cloaks, shields and rings are never
	// heavy, so they have no entry here.
	30: {5: 150, 20: 90, 21: 65, 19: 65, 25: 80, 26: 65},
}

// ArmorBands are the MaxArmor rows in order; each is the top tier of its band.
var ArmorBands = []int{4, 9, 14, 19, 25, 30}

// HeavyArmorTier is the first tier that can wear heavy armor.
const HeavyArmorTier = 10

var ArmorClass = map[int]string{
	0: "",
	1: "Light",
	2: "Medium",
	3: "Heavy",
	4: "Robe",
}

// SettableArmorClasses are the armor classes a builder may put on an item.
// Medium (2) is retired.
var SettableArmorClasses = []int{0, 1, 3, 4}

// UnclassedArmorTypes are the item types that are always unclassed: neck
// (cloaks), shields and rings fit everyone.
var UnclassedArmorTypes = []int{22, 23, 24}

// ArmorClassFitsType reports whether an item of this type may carry the
// given armor class.
func ArmorClassFitsType(itemType int, class int) bool {
	if class == 0 {
		return true
	}
	for _, t := range UnclassedArmorTypes {
		if t == itemType {
			return false
		}
	}
	return true
}

// HeavyArmorFloorBand is the last band whose armor can never be heavy: a
// heavy piece's armor must be above that band's cap, so heavy starts with
// tier 10-14 level armor.
const HeavyArmorFloorBand = 9

// HeavyArmorFits reports whether a piece of this type and armor value is high
// enough to be classed heavy.
func HeavyArmorFits(itemType int, armor int) bool {
	return armor > MaxArmor[HeavyArmorFloorBand][itemType]
}

// ArmorClassValid reports whether an item of this type and armor value may
// carry the armor class: none always may; otherwise it must be armor, not a
// cloak, shield or ring, and heavy needs armor above the t5-9 cap.
func ArmorClassValid(itemType int, armor int, class int) bool {
	if class == 0 {
		return true
	}
	isArmor := false
	for _, t := range ArmorTypes {
		if t == itemType {
			isArmor = true
		}
	}
	if !isArmor || !ArmorClassFitsType(itemType, class) {
		return false
	}
	return class != 3 || HeavyArmorFits(itemType, armor)
}

// ParseArmorClass reads an armor class given by name ("none", "light",
// "heavy", "robe") or number, and reports whether it is one a builder may set.
func ParseArmorClass(value string) (int, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	class, err := strconv.Atoi(value)
	if err != nil {
		class = -1
		for id, name := range ArmorClass {
			if (name != "" && value == strings.ToLower(name)) || (id == 0 && value == "none") {
				class = id
			}
		}
	}
	for _, ok := range SettableArmorClasses {
		if class == ok {
			return class, true
		}
	}
	return 0, false
}

// GearArmor is the armor a character gets from worn gear. gearArmor is the
// total armor of every worn piece and lightArmor the part of it on pieces
// that are light or lighter (light, robe, unclassed) and aren't shields.
// Everyone multiplies their gear by con. Fighters, thieves and rangers add dex
// on those pieces only, since moving well in light armor is a technique.
// Barbarians add max(dex, con) and their toughness on everything. Monks don't
// wear armor and are handled by MonkNaturalArmor.
func GearArmor(class, tier, con, dex, gearArmor, lightArmor int) int {
	bonus := float64(gearArmor) * float64(con) * ConArmorMod
	switch class {
	case FIGHTER, THIEF, RANGER:
		bonus += float64(lightArmor) * float64(dex) * DexArmorMod
	case BARBARIAN:
		extra := float64(max(dex, con)) * ConArmorMod
		if tier >= ToughnessTier {
			extra += float64(tier) * ToughnessPerTier
		}
		bonus += float64(gearArmor) * extra
	}
	return gearArmor + int(bonus)
}

// ArmorDescription names a piece of armor for players: "heavy body armor",
// or just "ring armor" when it is unclassed.
func ArmorDescription(itemType int, armorClass int) string {
	desc := ItemTypes[itemType] + " armor"
	if armorClass != 0 {
		desc = ArmorClassName(armorClass) + " " + desc
	}
	return desc
}

// ArmorClassName is the display name of an armor class; unclassed reads "none".
func ArmorClassName(class int) string {
	if class == 0 {
		return "none"
	}
	return strings.ToLower(ArmorClass[class])
}

// CheckArmor reports whether a piece of the given type, armor value and armor
// class is within the cap for a wearer of the given tier. Heavy pieces use the
// next band's cap. Tiers above the last player band (25) have no cap row.
func CheckArmor(aType, tier, val, armorClass int) bool {
	band := -1
	for i, top := range ArmorBands[:len(ArmorBands)-1] {
		if tier <= top {
			band = i
			break
		}
	}
	if band < 0 {
		return false
	}
	if armorClass == 3 {
		band++
	}
	return val <= MaxArmor[ArmorBands[band]][aType]
}

// MageArmorTypes are the only armor item types a mage can wear: body (robes,
// through ArmorRestrictions), neck (cloaks) and finger. Off hand items that
// aren't armor (lights, wands, devices...) are still fine; shields are not.
var MageArmorTypes = []int{5, 22, 24}

// MageCanWearArmorType reports whether a mage may equip an item of this type.
// Non-armor types always pass.
func MageCanWearArmorType(itemType int) bool {
	for _, t := range ArmorTypes {
		if t == itemType {
			for _, ok := range MageArmorTypes {
				if ok == itemType {
					return true
				}
			}
			return false
		}
	}
	return true
}

// ArmorRestrictions lists the armor classes each character class can wear.
// Robes (4) are open to every class. Heavy (3) is fighter and paladin only;
// everyone else who wears armor wears light (1). Medium (2) is unused, so it
// is left to the classes that can wear anything. Unclassed (0) pieces -
// necks, rings, shields and low tier gear - fit everyone.
var ArmorRestrictions = map[int][]int{
	FIGHTER:   {0, 1, 2, 3, 4},
	BARBARIAN: {0, 1, 4},
	THIEF:     {0, 1, 4},
	RANGER:    {0, 1, 4},
	MAGE:      {0, 4},
	CLERIC:    {0, 1, 4},
	PALADIN:   {0, 1, 2, 3, 4},
	BARD:      {0, 1, 4},
	MONK:      {0, 4},
}
