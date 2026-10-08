package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

// casterState builds a mage who knows vigor and summon, standing alone in a
// room, with the given input typed after the command.
func casterState(input ...string) *state {
	s := attackerState(config.MAGE, nil)
	s.actor.Name = "Caster"
	s.actor.Tier = 20
	s.actor.ParentId = -1
	s.actor.Spells = []string{"vigor", "summon"}
	s.actor.Int = objects.Meter{Max: 30, Current: 30}
	s.actor.Stam = objects.Meter{Max: 100, Current: 100}
	s.actor.Mana = objects.Meter{Max: 100, Current: 100}
	s.actor.Timers = map[string]time.Time{}
	s.actor.FlagProviders = map[string][]string{}
	s.actor.Effects = map[string]*objects.Effect{}
	s.actor.Inventory = &objects.ItemInventory{}
	s.where = &objects.Room{
		Flags: map[string]bool{},
		Chars: &objects.CharInventory{Contents: []*objects.Character{s.actor}},
		Mobs:  &objects.MobInventory{},
	}
	objects.Rooms[-1] = s.where
	s.msg.Actor = &message.Buffer{}
	s.input = input
	for _, word := range input {
		s.words = append(s.words, strings.ToUpper(word))
	}
	return s
}

// actorText returns what the command sent to the actor.
func actorText(s *state) string {
	var out bytes.Buffer
	s.msg.Actor.Deliver(&out)
	return out.String()
}

// A target name that matches nothing is refused: the spell must not land on
// the caster instead, and nothing is spent on it.
func TestCastRefusesUnknownTarget(t *testing.T) {
	for _, spell := range []string{"vigor", "summon"} {
		s := casterState(spell, "nobody")

		cast{}.process(s)

		if s.ok {
			t.Errorf("%s: cast on a missing target reported success", spell)
		}
		if s.actor.Mana.Current != 100 {
			t.Errorf("%s: mana = %d, charged for a refused cast", spell, s.actor.Mana.Current)
		}
		if _, ok := s.actor.Timers["cast"]; ok {
			t.Errorf("%s: cast timer started on a refused cast", spell)
		}
		if got := actorText(s); !strings.Contains(got, "'nobody'") {
			t.Errorf("%s: message %q does not name the missing target", spell, got)
		}
	}
}

// A device pointed at a name that matches nothing keeps its charge.
func TestUseRefusesUnknownTarget(t *testing.T) {
	s := casterState("wand", "nobody")
	wand := &objects.Item{Object: objects.Object{Name: "wand"}, Spell: "vigor", MaxUses: 3}
	s.actor.Inventory.Contents = append(s.actor.Inventory.Contents, wand)

	use{}.process(s)

	if s.ok {
		t.Error("use on a missing target reported success")
	}
	if wand.MaxUses != 3 {
		t.Errorf("charges = %d, spent on a refused use", wand.MaxUses)
	}
	if _, ok := s.actor.Timers["use"]; ok {
		t.Error("use timer started on a refused use")
	}
	if got := actorText(s); !strings.Contains(got, "'nobody'") {
		t.Errorf("message %q does not name the missing target", got)
	}
}

// Heal and restore each spend their own daily counter and stop at zero.
func TestSpendDailyChargeCountsDown(t *testing.T) {
	s := casterState()
	s.actor.ClassProps = map[string]int{"heals": 2, "restores": 1}

	if !spendDailyCharge(s, "heal") || !spendDailyCharge(s, "heal") {
		t.Fatal("caster with 2 heals should cast twice")
	}
	if spendDailyCharge(s, "heal") {
		t.Fatal("caster with no heals left should be refused")
	}
	if got := actorText(s); !strings.Contains(got, "cannot cast heal anymore today") {
		t.Errorf("refusal message %q", got)
	}
	if !spendDailyCharge(s, "restore") || spendDailyCharge(s, "restore") {
		t.Fatal("restore should spend its own counter")
	}
	if !spendDailyCharge(s, "vigor") {
		t.Fatal("vigor has no daily limit")
	}
}

// A self cast of heal with no charges left is refused before anything is spent.
func TestCastHealOnSelfWithoutChargesRefused(t *testing.T) {
	s := casterState("heal")
	s.actor.Class = config.CLERIC
	s.actor.Spells = []string{"heal"}
	s.actor.ClassProps = map[string]int{"heals": 0}

	cast{}.process(s)

	if s.actor.Mana.Current != 100 {
		t.Errorf("mana = %d, charged for a refused heal", s.actor.Mana.Current)
	}
	if got := actorText(s); !strings.Contains(got, "cannot cast heal anymore today") {
		t.Errorf("message %q", got)
	}
}

// Restore can never land on the caster, even with no target named.
func TestCastRestoreOnSelfRefused(t *testing.T) {
	s := casterState("restore")
	s.actor.Class = config.CLERIC
	s.actor.Spells = []string{"restore"}
	s.actor.ClassProps = map[string]int{"restores": 5}

	cast{}.process(s)

	if s.actor.ClassProps["restores"] != 5 {
		t.Errorf("restores = %d, charge spent on a refused self cast", s.actor.ClassProps["restores"])
	}
	if got := actorText(s); !strings.Contains(got, "only cast this spell on others") {
		t.Errorf("message %q", got)
	}
}

// Restore charges belong to the classes that can cast restore: cleric and bard.
func TestRefreshGivesRestoresToClericAndBard(t *testing.T) {
	for class, want := range map[int]int{config.CLERIC: 5, config.BARD: 5, config.PALADIN: 0} {
		c := &objects.Character{Class: class, ClassProps: map[string]int{}}
		c.Refresh()
		if got := c.ClassProps["restores"]; got != want {
			t.Errorf("class %d restores = %d, want %d", class, got, want)
		}
	}
}
