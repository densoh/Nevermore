package config

import "strings"

// MaxArmor Max Damage Reduction for level group and armor type
var MaxArmor = map[int]map[int]int{
	// Level int ref is cap
	4:  {5: 50, 20: 40, 21: 15, 19: 15, 25: 30, 22: 15, 26: 15, 23: 20, 24: 4},
	9:  {5: 70, 20: 50, 21: 25, 19: 25, 25: 40, 22: 25, 26: 25, 23: 40, 24: 6},
	14: {5: 90, 20: 60, 21: 35, 19: 35, 25: 50, 22: 35, 26: 35, 23: 60, 24: 8},
	19: {5: 110, 20: 70, 21: 45, 19: 45, 25: 60, 22: 45, 26: 45, 23: 80, 24: 10},
	25: {5: 130, 20: 80, 21: 55, 19: 55, 25: 70, 22: 55, 26: 55, 23: 100, 24: 12},
}

var ArmorClass = map[int]string{
	0: "",
	1: "Light",
	2: "Medium",
	3: "Heavy",
	4: "Robe",
}

// ArmorCap is the most armor an item of this type may have at the tier.
func ArmorCap(aType, tier int) int {
	for _, band := range []int{4, 9, 14, 19} {
		if tier <= band {
			return MaxArmor[band][aType]
		}
	}
	return MaxArmor[25][aType]
}

func CheckArmor(aType, tier, val int) bool {
	switch {
	case tier <= 4:
		if val <= MaxArmor[4][aType] {
			return true
		}
	case tier <= 9:
		if val <= MaxArmor[9][aType] {
			return true
		}
	case tier <= 14:
		if val <= MaxArmor[14][aType] {
			return true
		}
	case tier <= 19:
		if val <= MaxArmor[19][aType] {
			return true
		}
	case tier <= 25:
		if val <= MaxArmor[25][aType] {
			return true
		}
	}
	return false
}

// ArmorRestrictions are the armor classes each character class may wear.
// Every class can wear robes (4).
var ArmorRestrictions = map[int][]int{
	FIGHTER:   {0, 1, 2, 3, 4},
	BARBARIAN: {0, 1, 2, 3, 4},
	THIEF:     {0, 1, 2, 4},
	RANGER:    {0, 1, 2, 4},
	MAGE:      {0, 4},
	CLERIC:    {0, 4},
	PALADIN:   {0, 1, 2, 3, 4},
	BARD:      {0, 1, 4},
	MONK:      {0, 4},
}

// MageArmorTypes are the only armor slots a mage can fill: body (robes only),
// neck (cloaks) and rings. Non-armor off hand items are fine; shields are not.
var MageArmorTypes = []int{5, 22, 24}

// MageCanWear reports whether a mage may equip an item of this type and
// armor class. Items that aren't armor are always fine.
func MageCanWear(itemType int, armorClass int) bool {
	isArmor := false
	for _, t := range ArmorTypes {
		if t == itemType {
			isArmor = true
		}
	}
	if !isArmor {
		return true
	}
	if itemType == 5 {
		return armorClass == 4
	}
	for _, t := range MageArmorTypes {
		if t == itemType {
			return true
		}
	}
	return false
}

// ArmorDescription names a piece of armor for players: "heavy body armor",
// or just "ring armor" when it is unclassed.
func ArmorDescription(itemType int, armorClass int) string {
	desc := ItemTypes[itemType] + " armor"
	if armorClass != 0 {
		desc = strings.ToLower(ArmorClass[armorClass]) + " " + desc
	}
	return desc
}
