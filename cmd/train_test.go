package cmd

import (
	"strings"
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

// trainState builds a tier 1 sprite standing at a trainer with the experience
// and gold for tier 2, asking to train the two given stats.
func trainState(str, dex, con, intel, pie int, picks ...string) *state {
	s := attackerState(config.FIGHTER, nil)
	s.actor.Race = config.SPRITE
	s.actor.Tier = 1
	s.actor.Str = objects.Meter{Max: 30, Current: str}
	s.actor.Dex = objects.Meter{Max: 30, Current: dex}
	s.actor.Con = objects.Meter{Max: 30, Current: con}
	s.actor.Int = objects.Meter{Max: 30, Current: intel}
	s.actor.Pie = objects.Meter{Max: 30, Current: pie}
	s.actor.Experience.Value = config.TierExpLevels[2]
	s.actor.Gold.Value = config.GoldPerLevel[2]
	s.where = &objects.Room{Flags: map[string]bool{"train": true}}
	s.msg.Actor = &message.Buffer{}
	s.input = picks
	for _, pick := range picks {
		s.words = append(s.words, strings.ToUpper(pick))
	}
	return s
}

// A character rolled before a minimum was raised can still train, but not
// into stats other than the one that is short.
func TestTrainRefusesPicksThatSkipShortStat(t *testing.T) {
	s := trainState(9, 20, 10, 7, 4, "con", "int")

	train{}.process(s)

	if s.actor.Tier != 1 {
		t.Fatalf("tier = %d, trained without covering piety", s.actor.Tier)
	}
	if s.actor.Con.Current != 10 || s.actor.Int.Current != 7 || s.actor.Pie.Current != 4 {
		t.Errorf("stats changed on a refused train: con %d int %d pie %d", s.actor.Con.Current, s.actor.Int.Current, s.actor.Pie.Current)
	}
	if s.actor.Gold.Value != config.GoldPerLevel[2] {
		t.Errorf("gold charged on a refused train: %d left", s.actor.Gold.Value)
	}
}

// One point short owes one pick, the other is the player's to spend.
func TestTrainAcceptsPickCoveringShortStat(t *testing.T) {
	s := trainState(9, 20, 10, 7, 4, "con", "pie")

	train{}.process(s)

	if s.actor.Tier != 2 {
		t.Fatalf("tier = %d, want 2", s.actor.Tier)
	}
	if s.actor.Con.Current != 11 || s.actor.Pie.Current != 5 {
		t.Errorf("con %d pie %d, want 11 and 5", s.actor.Con.Current, s.actor.Pie.Current)
	}
}

// Short in two stats takes both picks.
func TestTrainNeedsBothPicksWhenTwoStatsShort(t *testing.T) {
	s := trainState(16, 20, 3, 7, 4, "pie", "int")
	train{}.process(s)
	if s.actor.Tier != 1 {
		t.Fatalf("tier = %d, trained with constitution still short", s.actor.Tier)
	}

	s = trainState(16, 20, 3, 7, 4, "con", "pie")
	train{}.process(s)
	if s.actor.Tier != 2 || s.actor.Con.Current != 4 || s.actor.Pie.Current != 5 {
		t.Errorf("tier %d con %d pie %d, want 2, 4 and 5", s.actor.Tier, s.actor.Con.Current, s.actor.Pie.Current)
	}
}

// With every stat on or over its minimum the picks are free.
func TestTrainUnrestrictedAtMinimums(t *testing.T) {
	s := trainState(9, 20, 8, 8, 5, "con", "int")

	train{}.process(s)

	if s.actor.Tier != 2 {
		t.Fatalf("tier = %d, want 2", s.actor.Tier)
	}
}

// Reroll still has to land on or over every minimum.
func TestRerollRefusesStatsUnderMinimum(t *testing.T) {
	s := trainState(9, 20, 10, 7, 4)
	// validateStats takes str, con, dex, int, pie
	if validateStats(s, 9, 10, 20, 7, 4) {
		t.Error("reroll accepted piety under the minimum")
	}
	if !validateStats(s, 8, 10, 20, 7, 5) {
		t.Error("reroll refused stats on the minimum")
	}
}
