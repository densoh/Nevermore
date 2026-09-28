package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

// recklessState builds a barbarian in the reckless stance holding main, with
// the given stamina. The stance is applied through the real effect so a drop
// clears the flag the way it does in play.
func recklessState(main *objects.Item, stam int) *state {
	s := attackerState(config.BARBARIAN, main)
	s.actor.FlagProviders = map[string][]string{}
	s.actor.Effects = map[string]*objects.Effect{}
	s.actor.Stam = objects.Meter{Max: 200, Current: stam}
	s.msg.Actor = &message.Buffer{}
	objects.Effects["reckless"](s.actor, s.actor, 0)
	return s
}

// The stance takes a stamina cut from every attack thrown while it is up.
func TestRecklessAttackCharges(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, 200)

	recklessAttack(s)

	if !s.actor.CheckFlag("reckless") {
		t.Fatal("stance dropped on a swing it could afford")
	}
	want := 200 - config.RecklessStamCost(200)
	if got := s.actor.Stam.Current; got != want {
		t.Errorf("stamina after one reckless attack = %d, want %d", got, want)
	}
}

// When the barbarian cannot cover the cost the stance drops instead, before
// the roll, so nothing is charged and the bonuses are not applied.
func TestRecklessAttackDropsWhenTired(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, config.RecklessStamCost(200)-1)
	before := s.actor.Stam.Current

	recklessAttack(s)

	if s.actor.CheckFlag("reckless") {
		t.Fatal("stance stayed up without the stamina to pay for it")
	}
	if s.actor.Stam.Current != before {
		t.Errorf("stamina changed on a dropped stance: %d -> %d", before, s.actor.Stam.Current)
	}
}

// A ranged weapon cannot be swung recklessly: the stance drops for free.
func TestRecklessAttackDropsForRangedWeapon(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 4}, 200)

	recklessAttack(s)

	if s.actor.CheckFlag("reckless") {
		t.Fatal("stance stayed up with a ranged weapon")
	}
	if s.actor.Stam.Current != 200 {
		t.Errorf("stamina charged for a ranged swing: %d", s.actor.Stam.Current)
	}
}

// Off the stance, attacks cost nothing.
func TestRecklessAttackFreeWhenOff(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.RemoveEffect("reckless")

	recklessAttack(s)

	if s.actor.Stam.Current != 200 {
		t.Errorf("stamina charged without the stance: %d", s.actor.Stam.Current)
	}
}

// The stance takes RecklessMissReduction points off the miss chance.
func TestDetermineMissChanceReckless(t *testing.T) {
	plain := attackerState(config.BARBARIAN, &objects.Item{ItemType: 1})
	plain.actor.Skills[1].Value = config.WeaponExpLevels[2]
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.Skills[1].Value = config.WeaponExpLevels[2]

	base := DetermineMissChance(plain, 0)
	got := DetermineMissChance(s, 0)
	if got != base-config.RecklessMissReduction {
		t.Errorf("reckless miss chance = %d, want %d (base %d)", got, base-config.RecklessMissReduction, base)
	}
}
