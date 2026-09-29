package config

import "testing"

func TestPaladinDivinityCap(t *testing.T) {
	if WeaponLevel(SkillExpLevels[10], PALADIN, DivinitySkill) != 9 {
		t.Error("paladin reaches divinity level 10")
	}
	if WeaponExpTitle(SkillExpLevels[10], PALADIN, DivinitySkill) != WeaponTitles[9] {
		t.Error("paladin given the grandmaster divinity title")
	}
	if WeaponExpNext(SkillExpLevels[9], PALADIN, DivinitySkill) != 0 {
		t.Error("paladin shown a next divinity level past 9")
	}
	if WeaponLevel(SkillExpLevels[10], CLERIC, DivinitySkill) != 10 {
		t.Error("cleric capped below divinity level 10")
	}
}

func TestWeaponGrandmasterClasses(t *testing.T) {
	reach := map[int]bool{FIGHTER: true}
	for class := range AvailableClasses {
		if class > MONK {
			continue
		}
		if got := WeaponLevel(SkillExpLevels[10], class, 1); (got == 10) != reach[class] {
			t.Errorf("class %d weapon level 10 reachable = %v", class, got == 10)
		}
	}
	if WeaponLevel(SkillExpLevels[10], RANGER, MissileSkill) != 10 {
		t.Error("ranger capped below missile level 10")
	}
	if WeaponLevel(SkillExpLevels[10], RANGER, 1) != 9 {
		t.Error("ranger reaches level 10 with a melee weapon")
	}
	if WeaponLevel(SkillExpLevels[10], MONK, HandSkill) != 10 {
		t.Error("monk capped below hand-to-hand level 10")
	}
}
