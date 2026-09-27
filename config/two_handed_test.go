package config

import "testing"

func TestTwoHandedCapBonus(t *testing.T) {
	cases := map[int]int{
		15:  5,  // 2.25 rounds to nothing, floor of one step
		45:  5,  // 6.75 rounds down
		50:  10, // 7.5 rounds up
		85:  15, // 12.75
		115: 15, // 17.25 rounds down
		120: 20, // 18
		140: 20, // 21
	}
	for cap, want := range cases {
		if got := TwoHandedCapBonus(cap); got != want {
			t.Errorf("cap %d bonus = %d, want %d", cap, got, want)
		}
	}
}

func TestWeaponDamageCap(t *testing.T) {
	if WeaponDamageCap(10, CLERIC, false) != 60 || WeaponDamageCap(10, CLERIC, true) != 70 {
		t.Error("tier 10 caps wrong")
	}
	if WeaponDamageCap(10, FIGHTER, false) != 65 || WeaponDamageCap(10, FIGHTER, true) != 75 {
		t.Error("fighter reads a tier ahead before the bonus")
	}
	if CanWield(10, CLERIC, 60, false) || !CanWield(10, CLERIC, 60, true) {
		t.Error("a max roll of 60 needs two hands at tier 10")
	}
	if CanWield(10, CLERIC, 70, true) {
		t.Error("two-handed cap is still a cap")
	}
}

func TestCanWieldTwoHanded(t *testing.T) {
	allowed := map[int]bool{FIGHTER: true, BARBARIAN: true, PALADIN: true}
	for class := FIGHTER; class <= MONK; class++ {
		if CanWieldTwoHanded(class) != allowed[class] {
			t.Errorf("class %d two-handed = %v", class, CanWieldTwoHanded(class))
		}
	}
}
