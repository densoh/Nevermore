package cmd

import (
	"testing"
	"time"

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
	s.actor.Tier = 15
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
	want := 200 - config.RecklessStamCost(200, 15)
	if got := s.actor.Stam.Current; got != want {
		t.Errorf("stamina after one reckless attack = %d, want %d", got, want)
	}
}

// When the barbarian cannot cover the cost the stance drops instead, before
// the roll, so nothing is charged and the bonuses are not applied.
func TestRecklessAttackDropsWhenTired(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, config.RecklessStamCost(200, 15)-1)
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
	plain.actor.Skills[1].Value = config.SkillExpLevels[2]
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.Skills[1].Value = config.SkillExpLevels[2]

	base := DetermineMissChance(plain, 0)
	got := DetermineMissChance(s, 0)
	if got != base-config.RecklessMissReduction {
		t.Errorf("reckless miss chance = %d, want %d (base %d)", got, base-config.RecklessMissReduction, base)
	}
}

// Circle neither pays for the stance nor gets its to-hit bonus.
func TestCircleMissChanceIgnoresReckless(t *testing.T) {
	plain := attackerState(config.BARBARIAN, &objects.Item{ItemType: 1})
	plain.actor.Skills[1].Value = config.SkillExpLevels[2]
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.Skills[1].Value = config.SkillExpLevels[2]

	if got, want := CircleMissChance(s, 0), DetermineMissChance(plain, 0); got != want {
		t.Errorf("reckless circle miss chance = %d, want %d", got, want)
	}
}

// The toggle ignores every other timer, so the stance drops mid-round or
// while stunned, and starts the stance cooldown.
func TestRecklessToggleIgnoresTimers(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.Timers = map[string]time.Time{
		"global": time.Now().Add(10 * time.Second),
		"combat": time.Now().Add(10 * time.Second),
		"stun":   time.Now().Add(10 * time.Second),
	}

	reckless{}.process(s)

	if s.actor.CheckFlag("reckless") {
		t.Fatal("stance stayed up though only the stance cooldown can block the toggle")
	}
	if ready, _ := s.actor.StanceReady(); ready {
		t.Error("toggling the stance did not start the stance cooldown")
	}
}

// A second toggle inside the cooldown is refused and leaves the stance alone.
func TestRecklessToggleBlockedByStanceCooldown(t *testing.T) {
	s := recklessState(&objects.Item{ItemType: 1}, 200)
	s.actor.Timers = map[string]time.Time{}
	s.actor.SetStanceTimer()

	reckless{}.process(s)

	if !s.actor.CheckFlag("reckless") {
		t.Fatal("stance dropped inside the stance cooldown")
	}
	if s.ok {
		t.Error("a refused toggle should not report success")
	}
}

// Flurry shares the same rule: only the stance cooldown gates the toggle.
func TestFlurryToggleStanceCooldown(t *testing.T) {
	s := attackerState(config.MONK, nil)
	s.actor.FlagProviders = map[string][]string{}
	s.actor.Effects = map[string]*objects.Effect{}
	s.actor.Timers = map[string]time.Time{
		"global": time.Now().Add(10 * time.Second),
		"combat": time.Now().Add(10 * time.Second),
		"stun":   time.Now().Add(10 * time.Second),
	}
	s.msg.Actor = &message.Buffer{}
	objects.Effects["flurry"](s.actor, s.actor, 0)

	flurry{}.process(s)
	if s.actor.CheckFlag("flurry") {
		t.Fatal("flurry stayed up though only the stance cooldown can block the toggle")
	}

	objects.Effects["flurry"](s.actor, s.actor, 0)
	flurry{}.process(s)
	if !s.actor.CheckFlag("flurry") {
		t.Fatal("flurry dropped inside the stance cooldown")
	}
}
