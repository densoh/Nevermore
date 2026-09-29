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
