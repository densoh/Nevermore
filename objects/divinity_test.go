package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func divinityCaster(class int) *Character {
	return &Character{
		Class:  class,
		Flags:  map[string]bool{},
		Skills: map[int]*Accumulator{config.DivinitySkill: {Value: config.WeaponExpLevels[5]}},
	}
}

func TestDivinityBonus(t *testing.T) {
	full := float64(config.HealingSkill[5])
	if got := divinityCaster(config.CLERIC).DivinityBonus(); got != full {
		t.Errorf("cleric bonus = %v, want %v", got, full)
	}
	paladin := divinityCaster(config.PALADIN)
	if got := paladin.DivinityBonus(); got != full*config.PaladinDivinityMod {
		t.Errorf("paladin bonus without seal = %v, want %v", got, full*config.PaladinDivinityMod)
	}
	paladin.Flags["seal-faith"] = true
	if got := paladin.DivinityBonus(); got != full {
		t.Errorf("paladin bonus with seal of faith = %v, want %v", got, full)
	}
}
