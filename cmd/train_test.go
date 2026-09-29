package cmd

import (
	"strings"
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

// trainState builds a tier 1 human standing at a trainer with the experience
// and gold for tier 2, asking to train the two given stats.
func trainState(str, dex, con, intel, pie int, picks ...string) *state {
	s := attackerState(config.FIGHTER, nil)
	s.actor.Race = config.HUMAN
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
	s := trainState(20, 10, 10, 5, 5, "dex", "con")

	train{}.process(s)

	if s.actor.Tier != 1 {
		t.Fatalf("tier = %d, trained without covering piety", s.actor.Tier)
	}
	if s.actor.Dex.Current != 10 || s.actor.Con.Current != 10 || s.actor.Pie.Current != 5 {
		t.Errorf("stats changed on a refused train: dex %d con %d pie %d", s.actor.Dex.Current, s.actor.Con.Current, s.actor.Pie.Current)
	}
	if s.actor.Gold.Value != config.GoldPerLevel[2] {
		t.Errorf("gold charged on a refused train: %d left", s.actor.Gold.Value)
	}
}

// One point short owes one pick, the other is the player's to spend.
func TestTrainAcceptsPickCoveringShortStat(t *testing.T) {
	s := trainState(20, 10, 10, 5, 5, "dex", "pie")

	train{}.process(s)

	if s.actor.Tier != 2 {
		t.Fatalf("tier = %d, want 2", s.actor.Tier)
	}
	if s.actor.Dex.Current != 11 || s.actor.Pie.Current != 6 {
		t.Errorf("dex %d pie %d, want 11 and 6", s.actor.Dex.Current, s.actor.Pie.Current)
	}
}

// Short in two stats takes both picks.
func TestTrainNeedsBothPicksWhenTwoStatsShort(t *testing.T) {
	s := trainState(20, 15, 5, 5, 5, "pie", "int")
	train{}.process(s)
	if s.actor.Tier != 1 {
		t.Fatalf("tier = %d, trained with constitution still short", s.actor.Tier)
	}

	s = trainState(20, 15, 5, 5, 5, "con", "pie")
	train{}.process(s)
	if s.actor.Tier != 2 || s.actor.Con.Current != 6 || s.actor.Pie.Current != 6 {
		t.Errorf("tier %d con %d pie %d, want 2, 6 and 6", s.actor.Tier, s.actor.Con.Current, s.actor.Pie.Current)
	}
}

// With every stat on or over its minimum the picks are free.
func TestTrainUnrestrictedAtMinimums(t *testing.T) {
	s := trainState(20, 8, 6, 10, 6, "dex", "int")

	train{}.process(s)

	if s.actor.Tier != 2 {
		t.Fatalf("tier = %d, want 2", s.actor.Tier)
	}
}

// Reroll still has to land on or over every minimum.
func TestRerollRefusesStatsUnderMinimum(t *testing.T) {
	s := trainState(20, 10, 10, 5, 5)
	if validateStats(s, 20, 10, 10, 5, 5) {
		t.Error("reroll accepted piety under the minimum")
	}
	if !validateStats(s, 19, 10, 10, 5, 6) {
		t.Error("reroll refused stats on the minimum")
	}
}
