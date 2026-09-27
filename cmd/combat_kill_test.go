package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
)

// attackerState builds the minimum actor DetermineMissChance needs: a class,
// a skill table, stats and a (possibly empty) equipment set.
func attackerState(class int, main *objects.Item) *state {
	skills := map[int]*objects.Accumulator{}
	for slot := 0; slot <= 11; slot++ {
		skills[slot] = &objects.Accumulator{}
	}
	return &state{actor: &objects.Character{
		Class:     class,
		Skills:    skills,
		Flags:     map[string]bool{},
		Modifiers: map[string]int{},
		Equipment: &objects.Equipment{Main: main},
	}}
}

// An unarmed monk must not dereference the empty main hand.  This crashed
// "kill" for every monk before the hand-to-hand skill was read first.
func TestDetermineMissChanceUnarmedMonk(t *testing.T) {
	s := attackerState(config.MONK, nil)
	s.actor.Skills[config.HandSkill].Value = config.WeaponExpLevels[3]

	got := DetermineMissChance(s, 0)
	want := config.WeaponMissChance(config.WeaponExpLevels[3])
	if got != want {
		t.Errorf("unarmed monk miss chance = %d, want %d (hand skill)", got, want)
	}
}

// A monk holding a weapon still fights with the hand-to-hand skill.
func TestDetermineMissChanceArmedMonkUsesHandSkill(t *testing.T) {
	s := attackerState(config.MONK, &objects.Item{ItemType: 0})
	s.actor.Skills[config.HandSkill].Value = config.WeaponExpLevels[5]
	s.actor.Skills[0].Value = 0

	got := DetermineMissChance(s, 0)
	want := config.WeaponMissChance(config.WeaponExpLevels[5])
	if got != want {
		t.Errorf("armed monk miss chance = %d, want %d (hand skill)", got, want)
	}
}

// Everyone else reads the wielded weapon's skill.
func TestDetermineMissChanceWeaponSkill(t *testing.T) {
	s := attackerState(config.FIGHTER, &objects.Item{ItemType: 1})
	s.actor.Skills[1].Value = config.WeaponExpLevels[4]
	s.actor.Skills[config.HandSkill].Value = config.WeaponExpLevels[9]

	got := DetermineMissChance(s, 0)
	want := config.WeaponMissChance(config.WeaponExpLevels[4])
	if got != want {
		t.Errorf("fighter miss chance = %d, want %d (weapon skill)", got, want)
	}
}
