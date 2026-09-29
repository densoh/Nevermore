package config

import "testing"

func TestMageReachesAscended(t *testing.T) {
	for slot := FirstElementSkill; slot <= LastElementSkill; slot++ {
		if WeaponLevel(SkillExpLevels[10], MAGE, slot) != 10 {
			t.Errorf("mage capped below affinity level 10 in slot %d", slot)
		}
		if WeaponExpNext(SkillExpLevels[9], MAGE, slot) != SkillExpLevels[10] {
			t.Errorf("mage next-level lookup stops at 9 in slot %d", slot)
		}
		if WeaponLevel(SkillExpLevels[10], CLERIC, slot) != 9 {
			t.Errorf("cleric unexpectedly reaches affinity level 10 in slot %d", slot)
		}
	}
	if WeaponLevel(SkillExpLevels[10], MAGE, 0) != 9 {
		t.Error("mage unexpectedly reaches weapon level 10")
	}
}

func TestSpellDamageBonus(t *testing.T) {
	// Tier 8 mage at Channeler: 8 from tier plus 20 from affinity.
	if got := SpellDamageBonus(8, 10, MAGE, SkillExpLevels[3], FirstElementSkill); got != 8*SpellTierDamagePercent+SpellDmgSkill[3] {
		t.Errorf("tier 8 channeler mage bonus = %d", got)
	}
	// Clerics get the tier bonus alone, whatever their affinity exp.
	if got := SpellDamageBonus(8, 10, CLERIC, SkillExpLevels[3], FirstElementSkill); got != 8*SpellTierDamagePercent {
		t.Errorf("tier 8 cleric bonus = %d", got)
	}
}

func TestDivineManaFromPiety(t *testing.T) {
	for _, class := range []int{CLERIC, PALADIN} {
		if CalcMana(14, 10, 25, class) != CalcMana(14, 25, 25, class) {
			t.Errorf("class %d mana pool still reads int", class)
		}
		want := 14*Classes[AvailableClasses[class]].Mana + int(14*25*ManaPerStatPerTier)
		if got := CalcMana(14, 10, 25, class); got != want {
			t.Errorf("class %d mana at tier 14 pie 25 = %d, want %d", class, got, want)
		}
	}
	if CalcMana(14, 25, 10, BARD) != CalcMana(14, 25, 25, BARD) {
		t.Error("bard mana pool should ignore piety")
	}
}

func TestSpellDamageBonusAddsInt(t *testing.T) {
	// Tier 25 int 40 mage at level 9: 25 + 30 + 70, added rather than compounded.
	want := 25*SpellTierDamagePercent + 30*SpellIntDamagePercent + SpellDmgSkill[9]
	if got := SpellDamageBonus(25, 40, MAGE, SkillExpLevels[9], FirstElementSkill); got != want {
		t.Errorf("tier 25 int 40 level 9 mage bonus = %d, want %d", got, want)
	}
	// Int below the baseline never subtracts.
	if SpellDamageBonus(5, 6, BARD, 0, FirstElementSkill) != 5*SpellTierDamagePercent {
		t.Error("low int reduced the spell bonus")
	}
}
