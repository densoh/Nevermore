package cmd

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

func TestMatchDevious(t *testing.T) {
	cases := map[string]string{
		"poison":          config.DeviousPoison,
		"poi":             config.DeviousPoison,
		"expose":          config.DeviousExpose,
		"expose-weakness": config.DeviousExpose,
		"silence":         "",
		"":                "",
	}
	for word, want := range cases {
		if got := matchDevious(word); got != want {
			t.Errorf("%q matched %q, want %q", word, got, want)
		}
	}
}

// A queued expose weakness waits on a mob that's already exposed, so that
// hit sneak attacks instead.
func TestDeviousForSkipsExposedMob(t *testing.T) {
	thief := sneakThief(12, 0)
	thief.DeviousStrike = config.DeviousExpose
	mob := sneakMob("Tank")
	mob.Flags = map[string]bool{}
	if got := deviousFor(thief, mob); got != config.DeviousExpose {
		t.Fatalf("fresh mob: %q, want expose", got)
	}
	mob.Expose()
	if got := deviousFor(thief, mob); got != "" {
		t.Errorf("exposed mob: %q, want sneak attack", got)
	}
	thief.DeviousStrike = config.DeviousPoison
	if got := deviousFor(thief, mob); got != config.DeviousPoison {
		t.Errorf("poison on an exposed mob: %q, want poison", got)
	}
}

// Expose weakness lands once, clears the queue, and a second one does
// nothing and stays queued.
func TestFireDeviousExpose(t *testing.T) {
	s := attackerState(config.THIEF, &objects.Item{ItemType: 0})
	s.msg.Actor = &message.Buffer{}
	mob := sneakMob("Tank")
	mob.Flags = map[string]bool{}
	mob.Armor = 50

	s.actor.DeviousStrike = config.DeviousExpose
	fireDevious(s, mob, config.DeviousExpose)
	if s.actor.DeviousStrike != "" {
		t.Error("strike still queued after landing")
	}
	if got := mob.EffectiveArmor(); got != 50-config.ExposeWeaknessArmor {
		t.Errorf("armor %d, want %d", got, 50-config.ExposeWeaknessArmor)
	}

	s.actor.DeviousStrike = config.DeviousExpose
	fireDevious(s, mob, config.DeviousExpose)
	if s.actor.DeviousStrike != config.DeviousExpose {
		t.Error("a second expose was spent though it can't stack")
	}
	if got := mob.EffectiveArmor(); got != 50-config.ExposeWeaknessArmor {
		t.Errorf("expose stacked: armor %d", got)
	}
}

// A mob the hit killed takes no strike, and it stays queued.
func TestFireDeviousKeepsStrikeOnKill(t *testing.T) {
	s := attackerState(config.THIEF, &objects.Item{ItemType: 0})
	s.msg.Actor = &message.Buffer{}
	mob := sneakMob("Tank")
	mob.Flags = map[string]bool{}
	mob.Stam.Current = 0
	s.actor.DeviousStrike = config.DeviousPoison
	fireDevious(s, mob, config.DeviousPoison)
	if s.actor.DeviousStrike != config.DeviousPoison {
		t.Error("strike spent on a dead mob")
	}
}

// Strikes must be learned from a scroll before they can be readied.
func TestKnowsDevious(t *testing.T) {
	s := attackerState(config.THIEF, &objects.Item{ItemType: 0})
	if knowsDevious(s, config.DeviousPoison) {
		t.Fatal("unlearned strike reported known")
	}
	s.actor.Spells = []string{"light", config.DeviousSpell(config.DeviousPoison)}
	if !knowsDevious(s, config.DeviousPoison) || knowsDevious(s, config.DeviousExpose) {
		t.Error("known strikes not read from the spell list")
	}
}

// Readying has its own one-second cooldown and nothing else: a stunned thief
// can still ready, but can't change strikes twice in a row.
func TestDeviousCooldown(t *testing.T) {
	s := attackerState(config.THIEF, &objects.Item{ItemType: 0})
	s.msg.Actor = &message.Buffer{}
	s.actor.Tier = 12
	s.actor.Spells = []string{config.DeviousSpell(config.DeviousPoison), config.DeviousSpell(config.DeviousExpose)}
	s.actor.Timers = map[string]time.Time{"stun": time.Now().Add(time.Minute), "global": time.Now().Add(time.Minute)}

	s.words = []string{"POISON"}
	devious{}.process(s)
	if s.actor.DeviousStrike != config.DeviousPoison {
		t.Fatalf("stunned thief couldn't ready: %q", s.actor.DeviousStrike)
	}
	s.words = []string{"EXPOSE"}
	devious{}.process(s)
	if s.actor.DeviousStrike != config.DeviousPoison {
		t.Error("changed strikes inside the cooldown")
	}
	s.actor.Timers["devious"] = time.Now().Add(-time.Second)
	devious{}.process(s)
	if s.actor.DeviousStrike != config.DeviousExpose {
		t.Errorf("couldn't change after the cooldown: %q", s.actor.DeviousStrike)
	}
}
